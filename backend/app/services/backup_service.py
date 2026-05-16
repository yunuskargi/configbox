import difflib
from datetime import datetime
from pathlib import Path

from app.config import APP_TIMEZONE
from sqlalchemy.orm import Session

from app.config import BACKUP_DIR
from app.models import Backup, Device
from app.services.fortigate import fetch_fortigate_config
from app.services.juniper import fetch_juniper_config
from app.services.cisco import fetch_cisco_config
from app.services.paloalto import fetch_paloalto_config
from app.services.mail_service import notify_backup, notify_config_change


def _get_previous_config(device_dir: Path):
    confs = sorted(device_dir.glob("*.conf"), key=lambda f: f.stat().st_mtime, reverse=True)
    if confs:
        return confs[0].read_text(encoding="utf-8")
    return None


def _detect_changes(old_config: str, new_config: str) -> list[str]:
    old_lines = old_config.splitlines(keepends=True)
    new_lines = new_config.splitlines(keepends=True)
    diff = list(difflib.unified_diff(old_lines, new_lines, fromfile="onceki", tofile="yeni", lineterm=""))
    return diff


def run_backup(device: Device, db: Session, triggered_by: str = "manual") -> dict:
    timestamp = datetime.now(APP_TIMEZONE).strftime("%Y-%m-%d_%H%M%S")
    device_dir = BACKUP_DIR / device.vendor / device.name
    device_dir.mkdir(parents=True, exist_ok=True)

    previous_config = _get_previous_config(device_dir)

    file_path = device_dir / f"{timestamp}.conf"

    try:
        if device.vendor == "fortigate":
            config = fetch_fortigate_config(device)
        elif device.vendor == "juniper":
            config = fetch_juniper_config(device)
        elif device.vendor == "cisco":
            config = fetch_cisco_config(device)
        elif device.vendor == "paloalto":
            config = fetch_paloalto_config(device)
        else:
            raise ValueError(f"Unsupported vendor: {device.vendor}")

        file_path.write_text(config, encoding="utf-8")
        file_size = file_path.stat().st_size

        config_changed = False
        diff_lines = []
        if previous_config is not None:
            diff_lines = _detect_changes(previous_config, config)
            config_changed = len(diff_lines) > 0

        backup = Backup(
            device_id=device.id,
            file_path=str(file_path),
            file_size=file_size,
            status="success",
            triggered_by=triggered_by,
        )
        db.add(backup)
        db.commit()
        try:
            location_name = device.location.name if device.location else None
            notify_backup(device.name, device.vendor, "success",
                          file_path=str(file_path), file_size=file_size,
                          location=location_name, vdom=device.vdom,
                          triggered_by=triggered_by)
            if config_changed:
                diff_text = "\n".join(diff_lines[:100])
                notify_config_change(device.name, device.vendor, diff_text,
                                     location=location_name, vdom=device.vdom)
        except Exception:
            pass
        return {"status": "success", "file_path": str(file_path), "file_size": file_size,
                "config_changed": config_changed}

    except Exception as e:
        backup = Backup(
            device_id=device.id,
            file_path=str(file_path),
            file_size=0,
            status="failed",
            error_message=str(e),
            triggered_by=triggered_by,
        )
        db.add(backup)
        db.commit()
        try:
            location_name = device.location.name if device.location else None
            notify_backup(device.name, device.vendor, "failed", error=str(e),
                          location=location_name, vdom=device.vdom,
                          triggered_by=triggered_by)
        except Exception:
            pass
        return {"status": "failed", "error": str(e)}

import csv
import io
from fastapi import APIRouter, Depends, HTTPException, Query, Request, UploadFile, File
from fastapi.responses import StreamingResponse
from sqlalchemy.orm import Session
from sqlalchemy import func

from app.database import get_db
from app.models import Device, Backup, User, Location
from app.schemas import DeviceCreate, DeviceUpdate, DeviceOut, ScheduleUpdate
from app.auth import get_current_user
from app.services.backup_service import run_backup
from app.services.fortigate import test_fortigate
from app.services.juniper import test_juniper
from app.services.cisco import test_cisco
from app.services.paloalto import test_paloalto
from app.services.audit_service import log_action
from app.crypto import encrypt, decrypt

ENCRYPTED_FIELDS = ("auth_token", "ssh_password", "enable_password")


class DecryptedDevice:
    def __init__(self, device):
        self._device = device
    def __getattr__(self, name):
        val = getattr(self._device, name)
        if name in ENCRYPTED_FIELDS and val:
            return decrypt(val)
        return val

router = APIRouter(prefix="/api/devices", tags=["devices"])


def enrich_device(device: Device, db: Session) -> dict:
    last = db.query(Backup.created_at).filter(Backup.device_id == device.id, Backup.status == "success").order_by(Backup.created_at.desc()).first()
    success_count = db.query(func.count(Backup.id)).filter(Backup.device_id == device.id, Backup.status == "success").scalar()
    failed_count = db.query(func.count(Backup.id)).filter(Backup.device_id == device.id, Backup.status == "failed").scalar()
    d = {c.name: getattr(device, c.name) for c in device.__table__.columns}
    d["last_backup"] = last[0] if last else None
    d["backup_count"] = success_count or 0
    d["failed_count"] = failed_count or 0
    d["location_name"] = device.location.name if device.location else None
    d["has_token"] = bool(device.auth_token)
    d["has_ssh_password"] = bool(device.ssh_password)
    d["has_enable_password"] = bool(device.enable_password)
    return d


@router.get("", response_model=list[DeviceOut])
def list_devices(db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    devices = db.query(Device).order_by(Device.name).all()
    return [enrich_device(d, db) for d in devices]


@router.post("", response_model=DeviceOut, status_code=201)
def create_device(body: DeviceCreate, request: Request, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if db.query(Device).filter(Device.name == body.name).first():
        raise HTTPException(status_code=400, detail="Device name already exists")
    data = body.model_dump()
    for field in ENCRYPTED_FIELDS:
        if data.get(field):
            data[field] = encrypt(data[field])
    device = Device(**data)
    db.add(device)
    db.commit()
    db.refresh(device)
    log_action(db, user, "create", "device", device.name, f"Vendor: {device.vendor}, IP: {device.ip_address}", request.client.host)
    return enrich_device(device, db)


@router.get("/{device_id}", response_model=DeviceOut)
def get_device(device_id: int, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    device = db.query(Device).filter(Device.id == device_id).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")
    return enrich_device(device, db)


@router.put("/{device_id}", response_model=DeviceOut)
def update_device(device_id: int, body: DeviceUpdate, request: Request, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    device = db.query(Device).filter(Device.id == device_id).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")
    changes = body.model_dump(exclude_unset=True)
    changed_fields = [k for k in changes if k not in ENCRYPTED_FIELDS]
    for field, value in changes.items():
        if field in ENCRYPTED_FIELDS and value:
            value = encrypt(value)
        setattr(device, field, value)
    db.commit()
    db.refresh(device)
    log_action(db, user, "update", "device", device.name, f"Değişen: {', '.join(changed_fields) or 'credentials'}", request.client.host)
    return enrich_device(device, db)


@router.delete("/{device_id}")
def delete_device(device_id: int, keep_backups: bool = Query(True), request: Request = None, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    device = db.query(Device).filter(Device.id == device_id).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")
    device_name = device.name
    if not keep_backups:
        import shutil
        from app.config import BACKUP_DIR
        device_dir = BACKUP_DIR / device.vendor / device.name
        if device_dir.exists():
            shutil.rmtree(device_dir)
    db.delete(device)
    db.commit()
    log_action(db, user, "delete", "device", device_name, f"Yedekler korundu: {'Evet' if keep_backups else 'Hayır'}", request.client.host if request else None)
    return {"message": "Device deleted"}


@router.post("/{device_id}/backup")
def trigger_backup(device_id: int, request: Request = None, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    device = db.query(Device).filter(Device.id == device_id).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")
    result = run_backup(DecryptedDevice(device), db, triggered_by="manual")
    log_action(db, user, "backup", "device", device.name, f"Sonuç: {result['status']}", request.client.host if request else None)
    return result


@router.post("/{device_id}/test")
def test_connection(device_id: int, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    device = db.query(Device).filter(Device.id == device_id).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")
    try:
        dd = DecryptedDevice(device)
        if device.vendor == "fortigate":
            test_fortigate(dd)
        elif device.vendor == "juniper":
            test_juniper(dd)
        elif device.vendor == "cisco":
            test_cisco(dd)
        elif device.vendor == "paloalto":
            test_paloalto(dd)
        return {"status": "success", "message": "Connection successful"}
    except Exception as e:
        return {"status": "failed", "message": str(e)}


@router.put("/{device_id}/schedule")
def set_schedule(device_id: int, body: ScheduleUpdate, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    device = db.query(Device).filter(Device.id == device_id).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")
    device.schedule_cron = body.cron
    db.commit()
    from app.services.scheduler_service import reschedule_device
    reschedule_device(device)
    return {"message": "Schedule updated", "cron": body.cron}


@router.delete("/{device_id}/schedule")
def remove_schedule(device_id: int, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    device = db.query(Device).filter(Device.id == device_id).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")
    device.schedule_cron = None
    db.commit()
    from app.services.scheduler_service import remove_device_schedule
    remove_device_schedule(device.id)
    return {"message": "Schedule removed"}


CSV_COLUMNS = ["name", "vendor", "ip_address", "port", "platform", "auth_token",
               "ssh_username", "ssh_password", "enable_password", "location", "vdom"]
VALID_VENDORS = {"fortigate", "juniper", "cisco", "paloalto"}
VALID_PLATFORMS = {"ios", "nxos", "asa"}


@router.get("/bulk/template")
def download_csv_template(user: User = Depends(get_current_user)):
    buf = io.StringIO()
    w = csv.writer(buf)
    w.writerow(CSV_COLUMNS)
    w.writerow(["FW-Istanbul", "fortigate", "10.0.1.1", "443", "", "api-key-here", "", "", "", "Istanbul DC", ""])
    w.writerow(["SW-Core", "cisco", "10.0.1.2", "22", "ios", "", "admin", "pass123", "enable123", "Ankara DC", ""])
    w.writerow(["JUN-Edge", "juniper", "10.0.1.3", "22", "", "", "admin", "pass456", "", "", ""])
    buf.seek(0)
    return StreamingResponse(
        iter([buf.getvalue()]),
        media_type="text/csv",
        headers={"Content-Disposition": "attachment; filename=confbox_devices_template.csv"},
    )


@router.post("/bulk/preview")
async def bulk_preview(file: UploadFile = File(...), db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if not file.filename.endswith(".csv"):
        raise HTTPException(status_code=400, detail="Sadece CSV dosyası yüklenebilir")
    content = (await file.read()).decode("utf-8-sig", errors="replace")
    reader = csv.DictReader(io.StringIO(content))
    existing_names = {d.name for d in db.query(Device.name).all()}
    location_map = {loc.name.lower(): loc.id for loc in db.query(Location).all()}
    rows = []
    for i, row in enumerate(reader):
        errors = []
        name = row.get("name", "").strip()
        vendor = row.get("vendor", "").strip().lower()
        ip = row.get("ip_address", "").strip()
        port_str = row.get("port", "").strip()
        platform = row.get("platform", "").strip().lower() or "ios"
        if not name:
            errors.append("Ad boş")
        elif name in existing_names:
            errors.append("Bu isimde cihaz var")
        if vendor not in VALID_VENDORS:
            errors.append(f"Geçersiz vendor: {vendor}")
        if not ip:
            errors.append("IP boş")
        port = 0
        try:
            port = int(port_str) if port_str else (443 if vendor == "fortigate" else 22)
        except ValueError:
            errors.append("Geçersiz port")
        if vendor == "cisco" and platform not in VALID_PLATFORMS:
            errors.append(f"Geçersiz platform: {platform}")
        loc_name = row.get("location", "").strip()
        loc_id = location_map.get(loc_name.lower()) if loc_name else None
        rows.append({
            "row": i + 1,
            "name": name, "vendor": vendor, "ip_address": ip, "port": port,
            "platform": platform,
            "auth_token": bool(row.get("auth_token", "").strip()),
            "ssh_username": row.get("ssh_username", "").strip(),
            "ssh_password": bool(row.get("ssh_password", "").strip()),
            "enable_password": bool(row.get("enable_password", "").strip()),
            "location": loc_name, "location_id": loc_id,
            "vdom": row.get("vdom", "").strip(),
            "errors": errors,
            "valid": len(errors) == 0,
        })
        existing_names.add(name)
    return {"rows": rows, "total": len(rows), "valid": sum(1 for r in rows if r["valid"]), "invalid": sum(1 for r in rows if not r["valid"])}


@router.post("/bulk/import")
async def bulk_import(file: UploadFile = File(...), request: Request = None, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if not file.filename.endswith(".csv"):
        raise HTTPException(status_code=400, detail="Sadece CSV dosyası yüklenebilir")
    content = (await file.read()).decode("utf-8-sig", errors="replace")
    reader = csv.DictReader(io.StringIO(content))
    existing_names = {d.name for d in db.query(Device.name).all()}
    location_map = {loc.name.lower(): loc.id for loc in db.query(Location).all()}
    created, skipped = 0, 0
    for row in reader:
        name = row.get("name", "").strip()
        vendor = row.get("vendor", "").strip().lower()
        ip = row.get("ip_address", "").strip()
        port_str = row.get("port", "").strip()
        platform = row.get("platform", "").strip().lower() or "ios"
        if not name or not ip or vendor not in VALID_VENDORS or name in existing_names:
            skipped += 1
            continue
        try:
            port = int(port_str) if port_str else (443 if vendor == "fortigate" else 22)
        except ValueError:
            skipped += 1
            continue
        loc_name = row.get("location", "").strip()
        loc_id = location_map.get(loc_name.lower()) if loc_name else None
        data = {
            "name": name, "vendor": vendor, "ip_address": ip, "port": port,
            "platform": platform, "location_id": loc_id,
            "vdom": row.get("vdom", "").strip() or None,
            "auth_token": row.get("auth_token", "").strip() or None,
            "ssh_username": row.get("ssh_username", "").strip() or None,
            "ssh_password": row.get("ssh_password", "").strip() or None,
            "enable_password": row.get("enable_password", "").strip() or None,
        }
        for field in ENCRYPTED_FIELDS:
            if data.get(field):
                data[field] = encrypt(data[field])
        device = Device(**data)
        db.add(device)
        existing_names.add(name)
        created += 1
    db.commit()
    log_action(db, user, "bulk_import", "device", f"{created} cihaz", f"Eklenen: {created}, Atlanan: {skipped}", request.client.host if request else None)
    return {"created": created, "skipped": skipped}

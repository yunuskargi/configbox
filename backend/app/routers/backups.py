import time
import hmac
import hashlib
from typing import Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from fastapi.responses import FileResponse
from sqlalchemy.orm import Session
from pathlib import Path

from app.database import get_db
from app.models import Backup, Device, User
from app.schemas import BackupOut
from app.auth import get_current_user, verify_password
from app.config import JWT_SECRET

router = APIRouter(prefix="/api/backups", tags=["backups"])

DOWNLOAD_TOKEN_TTL = 300


def _make_download_token(user_id: int, backup_id: int) -> str:
    ts = int(time.time())
    msg = f"{user_id}:{backup_id}:{ts}"
    sig = hmac.new(JWT_SECRET.encode(), msg.encode(), hashlib.sha256).hexdigest()[:32]
    return f"{msg}:{sig}"


def _verify_download_token(token: str, backup_id: int) -> bool:
    try:
        parts = token.rsplit(":", 1)
        if len(parts) != 2:
            return False
        msg, sig = parts
        uid, bid, ts = msg.split(":")
        if int(bid) != backup_id:
            return False
        if int(time.time()) - int(ts) > DOWNLOAD_TOKEN_TTL:
            return False
        expected = hmac.new(JWT_SECRET.encode(), msg.encode(), hashlib.sha256).hexdigest()[:32]
        return hmac.compare_digest(sig, expected)
    except Exception:
        return False


@router.get("", response_model=list[BackupOut])
def list_backups(
    device_id: Optional[int] = None,
    vendor: Optional[str] = None,
    status: Optional[str] = None,
    limit: int = Query(50, le=500),
    offset: int = 0,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    q = db.query(Backup).join(Device)
    if device_id:
        q = q.filter(Backup.device_id == device_id)
    if vendor:
        q = q.filter(Device.vendor == vendor)
    if status:
        q = q.filter(Backup.status == status)
    backups = q.order_by(Backup.created_at.desc()).offset(offset).limit(limit).all()
    results = []
    for b in backups:
        d = {c.name: getattr(b, c.name) for c in b.__table__.columns}
        d["device_name"] = b.device.name
        d["vendor"] = b.device.vendor
        d["location_name"] = b.device.location.name if b.device.location else None
        results.append(d)
    return results


from pydantic import BaseModel

class DownloadAuthRequest(BaseModel):
    password: str
    backup_id: int


@router.post("/authorize-download")
def authorize_download(body: DownloadAuthRequest, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if not verify_password(body.password, user.password_hash):
        raise HTTPException(status_code=403, detail="Şifre hatalı")
    backup = db.query(Backup).filter(Backup.id == body.backup_id).first()
    if not backup:
        raise HTTPException(status_code=404, detail="Backup not found")
    token = _make_download_token(user.id, body.backup_id)
    return {"download_token": token}


@router.get("/{backup_id}/download")
def download_backup(backup_id: int, token: str = Query(...), db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if not _verify_download_token(token, backup_id):
        raise HTTPException(status_code=403, detail="Geçersiz veya süresi dolmuş indirme tokeni")
    backup = db.query(Backup).filter(Backup.id == backup_id).first()
    if not backup:
        raise HTTPException(status_code=404, detail="Backup not found")
    file_path = Path(backup.file_path)
    if not file_path.exists():
        raise HTTPException(status_code=404, detail="Backup file not found on disk")
    return FileResponse(file_path, filename=file_path.name, media_type="application/octet-stream")


@router.get("/{backup_id}/content")
def get_backup_content(backup_id: int, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    backup = db.query(Backup).filter(Backup.id == backup_id).first()
    if not backup:
        raise HTTPException(status_code=404, detail="Backup not found")
    file_path = Path(backup.file_path)
    if not file_path.exists():
        raise HTTPException(status_code=404, detail="Backup file not found on disk")
    content = file_path.read_text(encoding="utf-8", errors="replace")
    return {"content": content, "file_path": str(file_path), "file_size": backup.file_size}


@router.get("/diff/{backup_id_a}/{backup_id_b}")
def diff_backups(backup_id_a: int, backup_id_b: int, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    ba = db.query(Backup).filter(Backup.id == backup_id_a).first()
    bb = db.query(Backup).filter(Backup.id == backup_id_b).first()
    if not ba or not bb:
        raise HTTPException(status_code=404, detail="Backup not found")
    fa, fb = Path(ba.file_path), Path(bb.file_path)
    if not fa.exists() or not fb.exists():
        raise HTTPException(status_code=404, detail="Backup file not found on disk")
    import difflib
    content_a = fa.read_text(encoding="utf-8", errors="replace").splitlines()
    content_b = fb.read_text(encoding="utf-8", errors="replace").splitlines()
    diff = list(difflib.unified_diff(content_a, content_b, fromfile=fa.name, tofile=fb.name, lineterm=""))
    stats = {"added": sum(1 for l in diff if l.startswith("+") and not l.startswith("+++")),
             "removed": sum(1 for l in diff if l.startswith("-") and not l.startswith("---"))}
    return {
        "diff": diff,
        "stats": stats,
        "backup_a": {"id": ba.id, "device_name": ba.device.name, "created_at": ba.created_at.isoformat()},
        "backup_b": {"id": bb.id, "device_name": bb.device.name, "created_at": bb.created_at.isoformat()},
    }


@router.delete("/{backup_id}")
def delete_backup(backup_id: int, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    backup = db.query(Backup).filter(Backup.id == backup_id).first()
    if not backup:
        raise HTTPException(status_code=404, detail="Backup not found")
    file_path = Path(backup.file_path)
    if file_path.exists():
        file_path.unlink()
    db.delete(backup)
    db.commit()
    return {"message": "Backup deleted"}

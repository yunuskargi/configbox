from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import Setting, User
from app.schemas import SettingsUpdate, SmtpSettings, NotifySettings
from app.auth import get_current_user
from app.services.mail_service import send_test_email

router = APIRouter(prefix="/api/settings", tags=["settings"])

DEFAULTS = {
    "backup_dir": "/data/backups",
    "retention_days": "90",
}


def get_setting(db: Session, key: str) -> str:
    row = db.query(Setting).filter(Setting.key == key).first()
    return row.value if row else DEFAULTS.get(key, "")


def set_setting(db: Session, key: str, value: str):
    row = db.query(Setting).filter(Setting.key == key).first()
    if row:
        row.value = value
    else:
        db.add(Setting(key=key, value=value))


@router.get("/branding")
def get_branding(db: Session = Depends(get_db)):
    return {
        "app_title": get_setting(db, "app_title") or "",
    }


@router.get("")
def get_settings(db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    return {
        "backup_dir": get_setting(db, "backup_dir"),
        "retention_days": int(get_setting(db, "retention_days") or 90),
        "app_title": get_setting(db, "app_title") or "",
    }


@router.put("")
def update_settings(body: SettingsUpdate, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin only")
    if body.backup_dir is not None:
        set_setting(db, "backup_dir", body.backup_dir)
    if body.retention_days is not None:
        set_setting(db, "retention_days", str(body.retention_days))
    if body.app_title is not None:
        set_setting(db, "app_title", body.app_title)
    db.commit()
    return {"message": "Settings updated"}


# --- SMTP ---
@router.get("/smtp")
def get_smtp(db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin only")
    keys = ["smtp_host", "smtp_port", "smtp_username", "smtp_password", "smtp_use_tls", "smtp_from_email", "smtp_from_name"]
    defaults = SmtpSettings()
    result = {}
    for k in keys:
        row = db.query(Setting).filter(Setting.key == k).first()
        result[k] = row.value if row else str(getattr(defaults, k))
    result["smtp_port"] = int(result["smtp_port"])
    result["smtp_use_tls"] = result["smtp_use_tls"].lower() == "true"
    return result


@router.put("/smtp")
def update_smtp(body: SmtpSettings, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin only")
    set_setting(db, "smtp_host", body.smtp_host)
    set_setting(db, "smtp_port", str(body.smtp_port))
    set_setting(db, "smtp_username", body.smtp_username)
    set_setting(db, "smtp_password", body.smtp_password)
    set_setting(db, "smtp_use_tls", str(body.smtp_use_tls).lower())
    set_setting(db, "smtp_from_email", body.smtp_from_email)
    set_setting(db, "smtp_from_name", body.smtp_from_name)
    db.commit()
    return {"message": "SMTP settings updated"}


@router.post("/smtp/test")
def test_smtp(db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin only")
    notify = get_notify(db, user)
    if not notify.get("notify_recipients"):
        raise HTTPException(status_code=400, detail="No recipients configured")
    try:
        send_test_email(notify["notify_recipients"])
        return {"message": "Test email sent"}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# --- Notify ---
@router.get("/notify")
def get_notify(db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin only")
    defaults = NotifySettings()
    keys = ["notify_on_success", "notify_on_failure", "notify_on_change", "notify_daily_summary", "notify_recipients"]
    result = {}
    for k in keys:
        row = db.query(Setting).filter(Setting.key == k).first()
        result[k] = row.value if row else str(getattr(defaults, k))
    for k in ["notify_on_success", "notify_on_failure", "notify_on_change", "notify_daily_summary"]:
        result[k] = result[k].lower() == "true"
    return result


@router.put("/notify")
def update_notify(body: NotifySettings, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin only")
    set_setting(db, "notify_on_success", str(body.notify_on_success).lower())
    set_setting(db, "notify_on_failure", str(body.notify_on_failure).lower())
    set_setting(db, "notify_on_change", str(body.notify_on_change).lower())
    set_setting(db, "notify_daily_summary", str(body.notify_daily_summary).lower())
    set_setting(db, "notify_recipients", body.notify_recipients)
    db.commit()
    return {"message": "Notification settings updated"}

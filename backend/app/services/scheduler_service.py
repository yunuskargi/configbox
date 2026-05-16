from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.cron import CronTrigger
from sqlalchemy.orm import Session

from app.database import SessionLocal
from app.models import Device
from app.services.backup_service import run_backup
from app.crypto import decrypt

ENCRYPTED_FIELDS = ("auth_token", "ssh_password", "enable_password")


class DecryptedDevice:
    def __init__(self, device):
        self._device = device
    def __getattr__(self, name):
        val = getattr(self._device, name)
        if name in ENCRYPTED_FIELDS and val:
            return decrypt(val)
        return val

scheduler = BackgroundScheduler()


def _backup_job(device_id: int):
    db = SessionLocal()
    try:
        device = db.query(Device).filter(Device.id == device_id).first()
        if device and device.is_active:
            run_backup(DecryptedDevice(device), db, triggered_by="scheduled")
    finally:
        db.close()


def schedule_device(device: Device):
    if not device.schedule_cron:
        return
    job_id = f"device_{device.id}"
    parts = device.schedule_cron.split()
    if len(parts) != 5:
        return
    trigger = CronTrigger(
        minute=parts[0], hour=parts[1], day=parts[2], month=parts[3], day_of_week=parts[4]
    )
    if scheduler.get_job(job_id):
        scheduler.reschedule_job(job_id, trigger=trigger)
    else:
        scheduler.add_job(_backup_job, trigger, args=[device.id], id=job_id, replace_existing=True)


def reschedule_device(device: Device):
    remove_device_schedule(device.id)
    schedule_device(device)


def remove_device_schedule(device_id: int):
    job_id = f"device_{device_id}"
    if scheduler.get_job(job_id):
        scheduler.remove_job(job_id)


def load_all_schedules():
    db = SessionLocal()
    try:
        devices = db.query(Device).filter(Device.schedule_cron.isnot(None), Device.is_active == True).all()
        for device in devices:
            schedule_device(device)
    finally:
        db.close()


def _daily_summary_job():
    from app.services.mail_service import send_daily_summary
    try:
        send_daily_summary()
    except Exception:
        pass


def start_scheduler():
    load_all_schedules()
    scheduler.add_job(
        _daily_summary_job,
        CronTrigger(hour=8, minute=0),
        id="daily_summary",
        replace_existing=True,
    )
    scheduler.start()


def shutdown_scheduler():
    scheduler.shutdown(wait=False)

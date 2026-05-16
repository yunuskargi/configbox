from datetime import datetime, timedelta
from typing import Optional
from app.config import APP_TIMEZONE
from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from sqlalchemy import func, cast, Date

from app.database import get_db
from app.models import Device, Backup, User, Location
from app.schemas import DashboardStats
from app.auth import get_current_user

router = APIRouter(prefix="/api/dashboard", tags=["dashboard"])


@router.get("/stats", response_model=DashboardStats)
def get_stats(db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    total_devices = db.query(func.count(Device.id)).scalar()
    active_devices = db.query(func.count(Device.id)).filter(Device.is_active == True).scalar()
    total_backups = db.query(func.count(Backup.id)).scalar()
    successful = db.query(func.count(Backup.id)).filter(Backup.status == "success").scalar()
    failed = db.query(func.count(Backup.id)).filter(Backup.status == "failed").scalar()
    scheduled = db.query(func.count(Device.id)).filter(Device.schedule_cron.isnot(None), Device.is_active == True).scalar()
    total_size = db.query(func.sum(Backup.file_size)).filter(Backup.status == "success").scalar() or 0

    today_start = datetime.now(APP_TIMEZONE).replace(hour=0, minute=0, second=0, microsecond=0)
    today_backups = db.query(func.count(Backup.id)).filter(Backup.created_at >= today_start).scalar()
    today_failed = db.query(func.count(Backup.id)).filter(Backup.created_at >= today_start, Backup.status == "failed").scalar()

    success_rate = round((successful / total_backups * 100), 1) if total_backups > 0 else 0.0

    vendor_dist = dict(
        db.query(Device.vendor, func.count(Device.id)).group_by(Device.vendor).all()
    )

    loc_rows = (
        db.query(Location.name, func.count(Device.id))
        .outerjoin(Device, Device.location_id == Location.id)
        .group_by(Location.name)
        .all()
    )
    location_dist = dict(loc_rows)

    recent = (
        db.query(Backup)
        .join(Device)
        .order_by(Backup.created_at.desc())
        .limit(10)
        .all()
    )
    activities = [
        {
            "id": b.id,
            "device_name": b.device.name,
            "vendor": b.device.vendor,
            "status": b.status,
            "triggered_by": b.triggered_by,
            "created_at": b.created_at.isoformat(),
            "error_message": b.error_message,
            "file_size": b.file_size,
        }
        for b in recent
    ]

    return DashboardStats(
        total_devices=total_devices,
        active_devices=active_devices,
        total_backups=total_backups,
        successful_backups=successful,
        failed_backups=failed,
        today_backups=today_backups,
        today_failed=today_failed,
        success_rate=success_rate,
        total_backup_size=total_size,
        vendor_distribution=vendor_dist,
        location_distribution=location_dist,
        recent_activities=activities,
        scheduled_devices=scheduled,
    )


@router.get("/trend")
def get_backup_trend(days: int = Query(30, le=365), db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    since = datetime.now(APP_TIMEZONE) - timedelta(days=days)
    rows = (
        db.query(
            func.date(Backup.created_at).label("date"),
            Backup.status,
            func.count(Backup.id).label("count"),
        )
        .filter(Backup.created_at >= since)
        .group_by(func.date(Backup.created_at), Backup.status)
        .all()
    )
    date_map = {}
    for row in rows:
        d = str(row.date)
        if d not in date_map:
            date_map[d] = {"date": d, "success": 0, "failed": 0}
        if row.status == "success":
            date_map[d]["success"] = row.count
        else:
            date_map[d]["failed"] = row.count
    current = since
    now = datetime.now(APP_TIMEZONE)
    result = []
    while current <= now:
        d = current.strftime("%Y-%m-%d")
        if d in date_map:
            result.append(date_map[d])
        else:
            result.append({"date": d, "success": 0, "failed": 0})
        current += timedelta(days=1)
    return result


@router.get("/size-trend")
def get_size_trend(days: int = Query(30, le=365), db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    since = datetime.now(APP_TIMEZONE) - timedelta(days=days)
    rows = (
        db.query(
            func.date(Backup.created_at).label("date"),
            func.sum(Backup.file_size).label("total_size"),
        )
        .filter(Backup.created_at >= since, Backup.status == "success")
        .group_by(func.date(Backup.created_at))
        .all()
    )
    date_map = {str(r.date): r.total_size or 0 for r in rows}
    current = since
    now = datetime.now(APP_TIMEZONE)
    result = []
    cumulative = 0
    while current <= now:
        d = current.strftime("%Y-%m-%d")
        day_size = date_map.get(d, 0)
        cumulative += day_size
        result.append({"date": d, "size": day_size, "cumulative": cumulative})
        current += timedelta(days=1)
    return result

from datetime import datetime
from sqlalchemy import Column, Integer, String, Boolean, DateTime, ForeignKey, Text
from sqlalchemy.orm import relationship

from app.database import Base
from app.config import APP_TIMEZONE


def now_local():
    return datetime.now(APP_TIMEZONE)


class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True)
    username = Column(String(100), unique=True, nullable=False, index=True)
    password_hash = Column(String(255), nullable=False)
    role = Column(String(20), nullable=False, default="admin")  # admin / backup_admin
    totp_secret = Column(String(32), nullable=True)
    totp_enabled = Column(Boolean, default=False)
    created_at = Column(DateTime, default=now_local)


class Location(Base):
    __tablename__ = "locations"

    id = Column(Integer, primary_key=True)
    name = Column(String(100), unique=True, nullable=False)
    description = Column(String(255), nullable=True)
    created_at = Column(DateTime, default=now_local)

    devices = relationship("Device", back_populates="location")


class Device(Base):
    __tablename__ = "devices"

    id = Column(Integer, primary_key=True)
    name = Column(String(100), nullable=False, unique=True)
    vendor = Column(String(20), nullable=False)
    ip_address = Column(String(45), nullable=False)
    port = Column(Integer, nullable=False)
    location_id = Column(Integer, ForeignKey("locations.id"), nullable=True)
    vdom = Column(String(100), nullable=True)
    auth_token = Column(String(255), nullable=True)
    ssh_username = Column(String(100), nullable=True)
    ssh_password = Column(String(255), nullable=True)
    enable_password = Column(String(255), nullable=True)
    platform = Column(String(20), nullable=True, default="ios")
    schedule_cron = Column(String(50), nullable=True)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=now_local)
    updated_at = Column(DateTime, default=now_local, onupdate=now_local)

    location = relationship("Location", back_populates="devices")
    backups = relationship("Backup", back_populates="device", cascade="all, delete-orphan")


class Backup(Base):
    __tablename__ = "backups"

    id = Column(Integer, primary_key=True)
    device_id = Column(Integer, ForeignKey("devices.id"), nullable=False)
    file_path = Column(String(500), nullable=False)
    file_size = Column(Integer, default=0)
    status = Column(String(20), nullable=False, default="success")
    error_message = Column(Text, nullable=True)
    triggered_by = Column(String(20), nullable=False, default="manual")
    created_at = Column(DateTime, default=now_local)

    device = relationship("Device", back_populates="backups")


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id = Column(Integer, primary_key=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=True)
    username = Column(String(100), nullable=False)
    action = Column(String(50), nullable=False)
    resource_type = Column(String(50), nullable=False)
    resource_name = Column(String(200), nullable=True)
    detail = Column(Text, nullable=True)
    ip_address = Column(String(45), nullable=True)
    created_at = Column(DateTime, default=now_local)


class Setting(Base):
    __tablename__ = "settings"

    key = Column(String(100), primary_key=True)
    value = Column(Text, nullable=True)

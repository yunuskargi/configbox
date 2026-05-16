from datetime import datetime
from typing import Optional
from pydantic import BaseModel


# --- Auth ---
class LoginRequest(BaseModel):
    username: str
    password: str
    totp_code: Optional[str] = None


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    role: str
    requires_2fa: bool = False


# --- User ---
class UserOut(BaseModel):
    id: int
    username: str
    role: str
    totp_enabled: bool = False
    created_at: datetime

    model_config = {"from_attributes": True}


class UserCreate(BaseModel):
    username: str
    password: str
    role: str = "backup_admin"  # admin / backup_admin


class UserUpdate(BaseModel):
    username: Optional[str] = None
    role: Optional[str] = None
    password: Optional[str] = None


class ChangePasswordRequest(BaseModel):
    current_password: str
    new_password: str


# --- Location ---
class LocationCreate(BaseModel):
    name: str
    description: Optional[str] = None


class LocationUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None


class LocationOut(BaseModel):
    id: int
    name: str
    description: Optional[str]
    device_count: int = 0
    created_at: datetime

    model_config = {"from_attributes": True}


# --- Device ---
class DeviceCreate(BaseModel):
    name: str
    vendor: str
    ip_address: str
    port: int = 443
    location_id: Optional[int] = None
    vdom: Optional[str] = None
    auth_token: Optional[str] = None
    ssh_username: Optional[str] = None
    ssh_password: Optional[str] = None
    enable_password: Optional[str] = None
    platform: Optional[str] = "ios"
    schedule_cron: Optional[str] = None


class DeviceUpdate(BaseModel):
    name: Optional[str] = None
    ip_address: Optional[str] = None
    port: Optional[int] = None
    location_id: Optional[int] = None
    vdom: Optional[str] = None
    auth_token: Optional[str] = None
    ssh_username: Optional[str] = None
    ssh_password: Optional[str] = None
    enable_password: Optional[str] = None
    platform: Optional[str] = None
    schedule_cron: Optional[str] = None
    is_active: Optional[bool] = None


class DeviceOut(BaseModel):
    id: int
    name: str
    vendor: str
    ip_address: str
    port: int
    location_id: Optional[int] = None
    location_name: Optional[str] = None
    vdom: Optional[str] = None
    platform: Optional[str] = None
    schedule_cron: Optional[str]
    is_active: bool
    created_at: datetime
    updated_at: datetime
    last_backup: Optional[datetime] = None
    backup_count: int = 0
    failed_count: int = 0
    has_token: bool = False
    has_ssh_password: bool = False
    has_enable_password: bool = False

    model_config = {"from_attributes": True}


class ScheduleUpdate(BaseModel):
    cron: str


# --- Backup ---
class BackupOut(BaseModel):
    id: int
    device_id: int
    device_name: str = ""
    vendor: str = ""
    location_name: Optional[str] = None
    file_path: str
    file_size: int
    status: str
    error_message: Optional[str]
    triggered_by: str
    created_at: datetime

    model_config = {"from_attributes": True}


# --- Dashboard ---
class DashboardStats(BaseModel):
    total_devices: int
    active_devices: int
    total_backups: int
    successful_backups: int
    failed_backups: int
    today_backups: int = 0
    today_failed: int = 0
    success_rate: float = 0.0
    total_backup_size: int = 0
    vendor_distribution: dict
    location_distribution: dict
    recent_activities: list
    scheduled_devices: int = 0


# --- Settings ---
class SettingsUpdate(BaseModel):
    backup_dir: Optional[str] = None
    retention_days: Optional[int] = None
    app_title: Optional[str] = None


# --- SMTP ---
class SmtpSettings(BaseModel):
    smtp_host: str = ""
    smtp_port: int = 587
    smtp_username: str = ""
    smtp_password: str = ""
    smtp_use_tls: bool = True
    smtp_from_email: str = ""
    smtp_from_name: str = "ConfBox"


# --- Audit Log ---
class AuditLogOut(BaseModel):
    id: int
    username: str
    action: str
    resource_type: str
    resource_name: Optional[str] = None
    detail: Optional[str] = None
    ip_address: Optional[str] = None
    created_at: datetime

    model_config = {"from_attributes": True}


class NotifySettings(BaseModel):
    notify_on_success: bool = False
    notify_on_failure: bool = True
    notify_on_change: bool = False
    notify_daily_summary: bool = False
    notify_recipients: str = ""

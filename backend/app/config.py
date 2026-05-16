import os
import secrets
from pathlib import Path
from datetime import timezone, timedelta

BASE_DIR = Path(__file__).resolve().parent.parent
BACKUP_DIR = Path(os.getenv("BACKUP_DIR", str(BASE_DIR.parent / "backups")))
DATABASE_URL = os.getenv("DATABASE_URL", f"sqlite:///{BASE_DIR}/confbox.db")

TZ_OFFSET = int(os.getenv("TZ_OFFSET", "3"))
APP_TIMEZONE = timezone(timedelta(hours=TZ_OFFSET))

_jwt_secret_file = BASE_DIR / ".jwt_secret"
def _get_jwt_secret():
    env_secret = os.getenv("JWT_SECRET")
    if env_secret and env_secret != "change-me-in-production":
        return env_secret
    if _jwt_secret_file.exists():
        return _jwt_secret_file.read_text().strip()
    generated = secrets.token_hex(32)
    _jwt_secret_file.write_text(generated)
    return generated

JWT_SECRET = _get_jwt_secret()
JWT_ALGORITHM = "HS256"
JWT_EXPIRE_MINUTES = int(os.getenv("JWT_EXPIRE_MINUTES", "480"))
DEFAULT_ADMIN_USER = os.getenv("DEFAULT_ADMIN_USER", "admin")
DEFAULT_ADMIN_PASS = os.getenv("DEFAULT_ADMIN_PASS", "admin")
CORS_ORIGINS = [o.strip() for o in os.getenv("CORS_ORIGINS", "").split(",") if o.strip()]

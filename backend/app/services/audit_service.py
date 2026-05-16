from sqlalchemy.orm import Session
from app.models import AuditLog, User


def log_action(db: Session, user: User, action: str, resource_type: str,
               resource_name: str = None, detail: str = None, ip_address: str = None):
    entry = AuditLog(
        user_id=user.id if user else None,
        username=user.username if user else "system",
        action=action,
        resource_type=resource_type,
        resource_name=resource_name,
        detail=detail,
        ip_address=ip_address,
    )
    db.add(entry)
    db.commit()

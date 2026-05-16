import io
import base64
import pyotp
import qrcode
from fastapi import APIRouter, Depends, HTTPException, Request, status
from sqlalchemy.orm import Session
from slowapi import Limiter
from slowapi.util import get_remote_address
from pydantic import BaseModel

from app.database import get_db
from app.models import User
from app.schemas import LoginRequest, TokenResponse, ChangePasswordRequest, UserOut
from app.auth import verify_password, create_token, hash_password, get_current_user, blacklist_token, validate_password
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from app.services.audit_service import log_action

bearer_scheme = HTTPBearer()

router = APIRouter(prefix="/api/auth", tags=["auth"])
limiter = Limiter(key_func=get_remote_address)


@router.post("/login", response_model=TokenResponse)
@limiter.limit("5/minute")
def login(request: Request, body: LoginRequest, db: Session = Depends(get_db)):
    user = db.query(User).filter(User.username == body.username).first()
    if not user or not verify_password(body.password, user.password_hash):
        log_action(db, None, "login_failed", "auth", body.username, "Başarısız giriş denemesi", request.client.host)
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid credentials")

    if user.totp_enabled and user.totp_secret:
        if not body.totp_code:
            return TokenResponse(access_token="", role=user.role, requires_2fa=True)
        totp = pyotp.TOTP(user.totp_secret)
        if not totp.verify(body.totp_code, valid_window=1):
            log_action(db, user, "login_failed", "auth", user.username, "Geçersiz 2FA kodu", request.client.host)
            raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Geçersiz 2FA kodu")

    token = create_token(user.username, user.role)
    log_action(db, user, "login", "auth", user.username, "Başarılı giriş", request.client.host)
    return TokenResponse(access_token=token, role=user.role)


@router.get("/me", response_model=UserOut)
def me(user: User = Depends(get_current_user)):
    return user


@router.post("/logout")
def logout(creds: HTTPAuthorizationCredentials = Depends(bearer_scheme)):
    blacklist_token(creds.credentials)
    return {"message": "Logged out"}


@router.post("/change-password")
def change_password(body: ChangePasswordRequest, user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    if not verify_password(body.current_password, user.password_hash):
        raise HTTPException(status_code=400, detail="Current password is incorrect")
    error = validate_password(body.new_password)
    if error:
        raise HTTPException(status_code=400, detail=error)
    user.password_hash = hash_password(body.new_password)
    db.commit()
    return {"message": "Password changed"}


@router.post("/2fa/setup")
def setup_2fa(user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    if user.totp_enabled:
        raise HTTPException(status_code=400, detail="2FA zaten aktif")
    secret = pyotp.random_base32()
    user.totp_secret = secret
    db.commit()
    totp = pyotp.TOTP(secret)
    uri = totp.provisioning_uri(name=user.username, issuer_name="ConfBox")
    img = qrcode.make(uri)
    buf = io.BytesIO()
    img.save(buf, format="PNG")
    qr_base64 = base64.b64encode(buf.getvalue()).decode()
    return {"secret": secret, "qr_code": f"data:image/png;base64,{qr_base64}"}


class Verify2FARequest(BaseModel):
    code: str


@router.post("/2fa/verify")
def verify_2fa(body: Verify2FARequest, user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    if not user.totp_secret:
        raise HTTPException(status_code=400, detail="Önce 2FA kurulumu yapın")
    totp = pyotp.TOTP(user.totp_secret)
    if not totp.verify(body.code, valid_window=1):
        raise HTTPException(status_code=400, detail="Geçersiz kod")
    user.totp_enabled = True
    db.commit()
    return {"message": "2FA aktif edildi"}


@router.post("/2fa/disable")
def disable_2fa(body: ChangePasswordRequest, user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    if not verify_password(body.current_password, user.password_hash):
        raise HTTPException(status_code=400, detail="Şifre hatalı")
    user.totp_secret = None
    user.totp_enabled = False
    db.commit()
    return {"message": "2FA devre dışı bırakıldı"}

from contextlib import asynccontextmanager
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from slowapi import Limiter
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded

from app.database import engine, Base, SessionLocal
from app.models import User
from app.auth import hash_password
from app.config import DEFAULT_ADMIN_USER, DEFAULT_ADMIN_PASS, BACKUP_DIR, CORS_ORIGINS
from app.services.scheduler_service import start_scheduler, shutdown_scheduler
from app.routers import auth, devices, backups, dashboard, settings, locations, users, audit

limiter = Limiter(key_func=get_remote_address)


def seed_admin():
    db = SessionLocal()
    try:
        if not db.query(User).filter(User.username == DEFAULT_ADMIN_USER).first():
            db.add(User(username=DEFAULT_ADMIN_USER, password_hash=hash_password(DEFAULT_ADMIN_PASS), role="admin"))
            db.commit()
    finally:
        db.close()


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    BACKUP_DIR.mkdir(parents=True, exist_ok=True)
    seed_admin()
    start_scheduler()
    yield
    shutdown_scheduler()


app = FastAPI(title="ConfBox", version="1.0.0", lifespan=lifespan)
app.state.limiter = limiter


@app.exception_handler(RateLimitExceeded)
async def rate_limit_handler(request: Request, exc: RateLimitExceeded):
    return JSONResponse(status_code=429, content={"detail": "Çok fazla istek. Lütfen bekleyin."})


app.add_middleware(
    CORSMiddleware,
    allow_origins=CORS_ORIGINS or ["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.middleware("http")
async def security_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["X-Frame-Options"] = "DENY"
    response.headers["X-XSS-Protection"] = "1; mode=block"
    response.headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
    return response


app.include_router(auth.router)
app.include_router(devices.router)
app.include_router(backups.router)
app.include_router(dashboard.router)
app.include_router(settings.router)
app.include_router(locations.router)
app.include_router(users.router)
app.include_router(audit.router)


@app.get("/api/health")
def health():
    return {"status": "ok"}

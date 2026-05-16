from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy import func

from app.database import get_db
from app.models import Location, Device, User
from app.schemas import LocationCreate, LocationUpdate, LocationOut
from app.auth import get_current_user

router = APIRouter(prefix="/api/locations", tags=["locations"])


def enrich_location(loc: Location, db: Session) -> dict:
    count = db.query(func.count(Device.id)).filter(Device.location_id == loc.id).scalar()
    return {
        "id": loc.id,
        "name": loc.name,
        "description": loc.description,
        "device_count": count or 0,
        "created_at": loc.created_at,
    }


@router.get("", response_model=list[LocationOut])
def list_locations(db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    locations = db.query(Location).order_by(Location.name).all()
    return [enrich_location(loc, db) for loc in locations]


@router.post("", response_model=LocationOut, status_code=201)
def create_location(body: LocationCreate, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if db.query(Location).filter(Location.name == body.name).first():
        raise HTTPException(status_code=400, detail="Location name already exists")
    loc = Location(**body.model_dump())
    db.add(loc)
    db.commit()
    db.refresh(loc)
    return enrich_location(loc, db)


@router.put("/{location_id}", response_model=LocationOut)
def update_location(location_id: int, body: LocationUpdate, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    loc = db.query(Location).filter(Location.id == location_id).first()
    if not loc:
        raise HTTPException(status_code=404, detail="Location not found")
    for field, value in body.model_dump(exclude_unset=True).items():
        setattr(loc, field, value)
    db.commit()
    db.refresh(loc)
    return enrich_location(loc, db)


@router.delete("/{location_id}")
def delete_location(location_id: int, db: Session = Depends(get_db), user: User = Depends(get_current_user)):
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin only")
    loc = db.query(Location).filter(Location.id == location_id).first()
    if not loc:
        raise HTTPException(status_code=404, detail="Location not found")
    db.query(Device).filter(Device.location_id == location_id).update({"location_id": None})
    db.delete(loc)
    db.commit()
    return {"message": "Location deleted"}

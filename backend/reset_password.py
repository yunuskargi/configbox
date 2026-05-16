#!/usr/bin/env python3
"""
ConfBox Admin Password Reset Tool
Usage: python reset_password.py [username] [new_password]
       python reset_password.py                          (interactive)
"""
import sys
from app.database import SessionLocal, engine, Base
from app.models import User
from app.auth import hash_password


def reset():
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()

    if len(sys.argv) == 3:
        username, new_pass = sys.argv[1], sys.argv[2]
    else:
        print("ConfBox - Şifre Sıfırlama")
        print("-" * 30)
        users = db.query(User).all()
        if not users:
            print("Hiç kullanıcı bulunamadı.")
            return
        print("Mevcut kullanıcılar:")
        for u in users:
            print(f"  - {u.username} ({u.role})")
        print()
        username = input("Kullanıcı adı: ").strip()
        new_pass = input("Yeni şifre: ").strip()

    if not username or not new_pass:
        print("Kullanıcı adı ve şifre boş olamaz.")
        sys.exit(1)

    user = db.query(User).filter(User.username == username).first()
    if not user:
        print(f"Kullanıcı bulunamadı: {username}")
        sys.exit(1)

    user.password_hash = hash_password(new_pass)
    db.commit()
    print(f"'{username}' kullanıcısının şifresi başarıyla sıfırlandı.")
    db.close()


if __name__ == "__main__":
    reset()

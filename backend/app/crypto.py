import base64
import hashlib
import os
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.primitives import padding

from app.config import JWT_SECRET

_KEY = hashlib.sha256(JWT_SECRET.encode()).digest()


def encrypt(plaintext: str) -> str:
    if not plaintext:
        return plaintext
    iv = os.urandom(16)
    padder = padding.PKCS7(128).padder()
    padded = padder.update(plaintext.encode("utf-8")) + padder.finalize()
    cipher = Cipher(algorithms.AES(_KEY), modes.CBC(iv))
    enc = cipher.encryptor()
    ct = enc.update(padded) + enc.finalize()
    return base64.b64encode(iv + ct).decode("ascii")


def decrypt(ciphertext: str) -> str:
    if not ciphertext:
        return ciphertext
    try:
        raw = base64.b64decode(ciphertext)
    except Exception:
        return ciphertext
    if len(raw) < 32:
        return ciphertext
    iv = raw[:16]
    ct = raw[16:]
    try:
        cipher = Cipher(algorithms.AES(_KEY), modes.CBC(iv))
        dec = cipher.decryptor()
        padded = dec.update(ct) + dec.finalize()
        unpadder = padding.PKCS7(128).unpadder()
        return (unpadder.update(padded) + unpadder.finalize()).decode("utf-8")
    except Exception:
        return ciphertext

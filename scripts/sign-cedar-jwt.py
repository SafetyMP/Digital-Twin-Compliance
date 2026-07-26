#!/usr/bin/env python3
"""Sign HS256 JWTs for cedar-service (matches services/cedar-service/internal/auth/jwt.go)."""
from __future__ import annotations

import base64
import hashlib
import hmac
import json
import os
import sys
import time


def b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode().rstrip("=")


def sign(secret: str, sub: str, roles: list[str], ttl_seconds: int = 3600) -> str:
    now = int(time.time())
    claims: dict = {
        "sub": sub,
        "roles": roles,
        "iat": now,
        "exp": now + ttl_seconds,
    }
    iss = os.environ.get("CEDAR_SERVICE_JWT_ISS", "").strip()
    aud = os.environ.get("CEDAR_SERVICE_JWT_AUD", "").strip()
    if iss:
        claims["iss"] = iss
    if aud:
        claims["aud"] = aud
    header = b64url(b'{"alg":"HS256","typ":"JWT"}')
    payload = b64url(json.dumps(claims, separators=(",", ":")).encode())
    sig_input = f"{header}.{payload}"
    sig = b64url(hmac.new(secret.encode(), sig_input.encode(), hashlib.sha256).digest())
    return f"{sig_input}.{sig}"


def main() -> None:
    if len(sys.argv) < 2:
        print("usage: sign-cedar-jwt.py <sub> [role1,role2,...] [ttl_seconds]", file=sys.stderr)
        sys.exit(2)
    secret = os.environ.get("CEDAR_SERVICE_JWT_SECRET", "").strip()
    if not secret:
        print("CEDAR_SERVICE_JWT_SECRET is required", file=sys.stderr)
        sys.exit(1)
    sub = sys.argv[1]
    roles = [r.strip() for r in sys.argv[2].split(",") if r.strip()] if len(sys.argv) > 2 else []
    ttl = 3600
    if len(sys.argv) > 3:
        ttl = int(sys.argv[3])
    print(sign(secret, sub, roles, ttl))


if __name__ == "__main__":
    main()

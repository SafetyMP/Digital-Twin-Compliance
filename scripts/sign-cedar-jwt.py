#!/usr/bin/env python3
"""Sign HS256 JWTs for cedar-service (matches services/cedar-service/internal/auth/jwt.go)."""
from __future__ import annotations

import base64
import hashlib
import hmac
import json
import os
import sys


def b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode().rstrip("=")


def sign(secret: str, sub: str, roles: list[str]) -> str:
    header = b64url(b'{"alg":"HS256","typ":"JWT"}')
    payload = b64url(json.dumps({"sub": sub, "roles": roles}, separators=(",", ":")).encode())
    sig_input = f"{header}.{payload}"
    sig = b64url(hmac.new(secret.encode(), sig_input.encode(), hashlib.sha256).digest())
    return f"{sig_input}.{sig}"


def main() -> None:
    if len(sys.argv) < 2:
        print("usage: sign-cedar-jwt.py <sub> [role1,role2,...]", file=sys.stderr)
        sys.exit(2)
    secret = os.environ.get("CEDAR_SERVICE_JWT_SECRET", "").strip()
    if not secret:
        print("CEDAR_SERVICE_JWT_SECRET is required", file=sys.stderr)
        sys.exit(1)
    sub = sys.argv[1]
    roles = [r.strip() for r in sys.argv[2].split(",") if r.strip()] if len(sys.argv) > 2 else []
    print(sign(secret, sub, roles))


if __name__ == "__main__":
    main()

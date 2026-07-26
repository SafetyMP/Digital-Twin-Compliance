from __future__ import annotations

import os
import time
from typing import Any
from urllib.parse import urljoin

import httpx
import jwt
from fastapi import FastAPI, HTTPException, Request, Response
from jwt import PyJWKClient
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

app = FastAPI(title="OIDC Edge", version="0.1.0", docs_url=None, redoc_url=None)


def _setup_otel() -> None:
    endpoint = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT")
    if not endpoint:
        return
    resource = Resource.create({"service.name": os.environ.get("OTEL_SERVICE_NAME", "oidc-edge")})
    provider = TracerProvider(resource=resource)
    provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=endpoint, insecure=True)))
    trace.set_tracer_provider(provider)
    FastAPIInstrumentor.instrument_app(app)
    HTTPXClientInstrumentor().instrument()


_setup_otel()

JWKS_URL = os.environ.get(
    "OIDC_JWKS_URL",
    "http://keycloak:8080/realms/digital-twin/protocol/openid-connect/certs",
)
ISSUERS = [
    s.strip()
    for s in os.environ.get(
        "OIDC_ISSUERS",
        "http://localhost:8088/realms/digital-twin,http://keycloak:8080/realms/digital-twin",
    ).split(",")
    if s.strip()
]
ROUTES = {
    "/state/": os.environ.get("UPSTREAM_STATE", "http://state-service:8080/"),
    "/alert/": os.environ.get("UPSTREAM_ALERT", "http://alert-service:8085/"),
    "/audit/": os.environ.get("UPSTREAM_AUDIT", "http://audit-service:8090/"),
    "/graph/": os.environ.get("UPSTREAM_GRAPH", "http://graph-service:8093/"),
    "/simulation/": os.environ.get("UPSTREAM_SIMULATION", "http://simulation-service:8094/"),
    "/reporting/": os.environ.get("UPSTREAM_REPORTING", "http://reporting-service:8095/"),
}

_jwks: PyJWKClient | None = None
_jwks_loaded_at = 0.0


def jwks_client() -> PyJWKClient:
    global _jwks, _jwks_loaded_at
    now = time.time()
    if _jwks is None or now - _jwks_loaded_at > 300:
        _jwks = PyJWKClient(JWKS_URL, cache_keys=True)
        _jwks_loaded_at = now
    return _jwks


def verify_bearer(auth_header: str | None) -> dict[str, Any]:
    if not auth_header or not auth_header.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="missing bearer token")
    token = auth_header[len("Bearer ") :].strip()
    if not token:
        raise HTTPException(status_code=401, detail="empty bearer token")
    try:
        key = jwks_client().get_signing_key_from_jwt(token)
        last_err: Exception | None = None
        for issuer in ISSUERS:
            try:
                return jwt.decode(
                    token,
                    key.key,
                    algorithms=["RS256"],
                    issuer=issuer,
                    options={"verify_aud": False},
                )
            except Exception as exc:  # noqa: BLE001
                last_err = exc
                continue
        raise last_err or RuntimeError("issuer validation failed")
    except HTTPException:
        raise
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=401, detail=f"invalid token: {exc}") from exc


@app.get("/api/v1/health")
async def health() -> dict[str, str]:
    return {"status": "ok", "role": "oidc-edge"}


@app.api_route("/{full_path:path}", methods=["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"])
async def proxy(full_path: str, request: Request) -> Response:
    path = "/" + full_path
    if path in ("/", "/favicon.ico"):
        raise HTTPException(status_code=404, detail="not found")

    upstream_base = None
    remainder = ""
    for prefix, base in ROUTES.items():
        if path.startswith(prefix):
            upstream_base = base
            remainder = path[len(prefix) :]
            break
    if upstream_base is None:
        raise HTTPException(status_code=404, detail="unknown route prefix")

    require_auth = not remainder.rstrip("/").endswith("api/v1/health")
    principal = "anonymous"
    roles: list[str] = []
    if require_auth:
        claims = verify_bearer(request.headers.get("Authorization"))
        principal = str(claims.get("preferred_username") or claims.get("sub") or "oidc-user")
        roles = list(claims.get("realm_access", {}).get("roles", []) or [])

    url = urljoin(upstream_base, remainder)
    if request.url.query:
        url = f"{url}?{request.url.query}"

    headers = {
        k: v
        for k, v in request.headers.items()
        if k.lower()
        not in {
            "host",
            "content-length",
            "transfer-encoding",
            "x-principal",
            "connection",
        }
    }
    if require_auth:
        headers["X-Principal"] = principal
        headers["X-Roles"] = ",".join(roles)

    body = await request.body()
    async with httpx.AsyncClient(timeout=60.0) as client:
        upstream = await client.request(request.method, url, content=body, headers=headers)
    return Response(
        content=upstream.content,
        status_code=upstream.status_code,
        media_type=upstream.headers.get("content-type"),
    )

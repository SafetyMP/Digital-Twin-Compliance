import { NextRequest, NextResponse } from "next/server";

const AUDIT_API = process.env.AUDIT_SERVICE_URL || "http://localhost:8090";

export async function GET(request: NextRequest) {
  const url = new URL(request.url);
  const target = `${AUDIT_API}/api/v1/audit/entries?${url.searchParams.toString()}`;
  const res = await fetch(target, { cache: "no-store" });
  const body = await res.text();
  return new NextResponse(body, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

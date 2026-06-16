import { NextRequest, NextResponse } from "next/server";

const AUDIT_API = process.env.AUDIT_SERVICE_URL || "http://localhost:8090";

export async function GET(
  _request: NextRequest,
  { params }: { params: { entryId: string } }
) {
  const res = await fetch(`${AUDIT_API}/api/v1/audit/entries/${params.entryId}`, {
    cache: "no-store",
  });
  const body = await res.text();
  return new NextResponse(body, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

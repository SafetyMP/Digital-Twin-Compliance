import { NextRequest, NextResponse } from "next/server";

const AUDIT_API = process.env.AUDIT_SERVICE_URL || "http://localhost:8090";

export async function GET(
  _request: NextRequest,
  context: { params: Promise<{ entryId: string }> }
) {
  const { entryId } = await context.params;
  const res = await fetch(`${AUDIT_API}/api/v1/audit/entries/${entryId}`, {
    cache: "no-store",
  });
  const body = await res.text();
  return new NextResponse(body, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

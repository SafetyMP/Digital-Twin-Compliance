import { NextRequest, NextResponse } from "next/server";

const ALERT_API = process.env.ALERT_SERVICE_URL || "http://localhost:8085";

export async function GET(
  _request: NextRequest,
  context: { params: Promise<{ alertId: string }> }
) {
  const { alertId } = await context.params;
  const res = await fetch(`${ALERT_API}/api/v1/alerts/${alertId}`, {
    cache: "no-store",
  });
  const body = await res.text();
  return new NextResponse(body, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

import { NextRequest, NextResponse } from "next/server";

const ALERT_API = process.env.ALERT_SERVICE_URL || "http://localhost:8085";

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ alertId: string }> }
) {
  const { alertId } = await context.params;
  const body = await request.text();
  const res = await fetch(`${ALERT_API}/api/v1/alerts/${alertId}/acknowledge`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
    cache: "no-store",
  });
  const responseBody = await res.text();
  return new NextResponse(responseBody, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

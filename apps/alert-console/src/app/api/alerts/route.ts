import { NextRequest, NextResponse } from "next/server";

const ALERT_API = process.env.ALERT_SERVICE_URL || "http://localhost:8085";

export async function GET(request: NextRequest) {
  const url = new URL(request.url);
  const target = `${ALERT_API}/api/v1/alerts?${url.searchParams.toString()}`;
  const res = await fetch(target, { cache: "no-store" });
  const body = await res.text();
  return new NextResponse(body, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

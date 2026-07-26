import { NextRequest, NextResponse } from "next/server";

const API = process.env.REPORTING_SERVICE_URL || "http://localhost:8095";

export async function POST(request: NextRequest) {
  const body = await request.text();
  const res = await fetch(`${API}/api/v1/reports`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
    cache: "no-store",
  });
  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

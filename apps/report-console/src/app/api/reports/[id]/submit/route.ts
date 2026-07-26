import { NextRequest, NextResponse } from "next/server";

const API = process.env.REPORTING_SERVICE_URL || "http://localhost:8095";

export async function POST(
  _request: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const res = await fetch(`${API}/api/v1/reports/${id}/submit`, {
    method: "POST",
    cache: "no-store",
  });
  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

import { NextRequest, NextResponse } from "next/server";

const GRAPH_API = process.env.GRAPH_SERVICE_URL || "http://localhost:8093";

export async function GET(request: NextRequest) {
  const url = new URL(request.url);
  const target = `${GRAPH_API}/api/v1/graph/summary?${url.searchParams.toString()}`;
  const res = await fetch(target, { cache: "no-store" });
  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

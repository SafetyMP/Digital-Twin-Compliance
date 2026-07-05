import { NextRequest, NextResponse } from "next/server";

const SIM_API = process.env.SIMULATION_SERVICE_URL || "http://localhost:8094";

export async function POST(request: NextRequest) {
  const body = await request.text();
  const res = await fetch(`${SIM_API}/api/v1/simulations/run`, {
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

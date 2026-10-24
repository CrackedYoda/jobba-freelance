import { NextResponse } from "next/server";

import { getServerEnv } from "@/lib/config/env";

const LOCAL_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);

/** Origins allowed to call mutating routes: the app URL, ALLOWED_ORIGINS, and any local host when running locally. */
export function isAllowedOrigin(origin: string): boolean {
  const env = getServerEnv();
  let parsed: URL;
  try { parsed = new URL(origin); } catch { return false; }
  const allowed = new Set([new URL(env.NEXT_PUBLIC_APP_URL).origin, ...(process.env.ALLOWED_ORIGINS ?? "").split(",").map((value) => value.trim()).filter(Boolean)]);
  if (allowed.has(parsed.origin)) return true;
  const runningLocally = env.NODE_ENV !== "production" || LOCAL_HOSTS.has(new URL(env.NEXT_PUBLIC_APP_URL).hostname);
  return runningLocally && LOCAL_HOSTS.has(parsed.hostname);
}

export function assertSameOrigin(request: Request): NextResponse | undefined {
  const origin = request.headers.get("origin");
  if (!origin) return;
  if (!isAllowedOrigin(origin)) {
    return NextResponse.json({ error: { code: "ORIGIN_MISMATCH", message: `Sign-in from ${origin} is not allowed. Open the site at ${new URL(getServerEnv().NEXT_PUBLIC_APP_URL).origin}.` } }, { status: 403 });
  }
}

/** The origin the user is actually on, used for the SIWS domain and URI so the wallet shows the right site. */
export function requestOrigin(request: Request): URL {
  const origin = request.headers.get("origin");
  return new URL(origin && isAllowedOrigin(origin) ? origin : getServerEnv().NEXT_PUBLIC_APP_URL);
}

export function invalidRequest(message = "Invalid request."): NextResponse {
  return NextResponse.json({ error: { code: "INVALID_REQUEST", message } }, { status: 400 });
}

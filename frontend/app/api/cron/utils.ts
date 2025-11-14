import { NextRequest } from "next/server";

/**
 * Verify that the request is from Vercel Cron
 * Vercel sends a special header or you can use a secret
 */
export function verifyCronRequest(request: NextRequest): boolean {
  // Option 1: Check for Vercel's cron header (if available)
  const cronHeader = request.headers.get("x-vercel-cron");
  if (cronHeader === "1") {
    return true;
  }

  // Option 2: Check for authorization header with CRON_SECRET
  const authHeader = request.headers.get("authorization");
  const cronSecret = process.env.CRON_SECRET;
  
  if (cronSecret && authHeader === `Bearer ${cronSecret}`) {
    return true;
  }

  // Option 3: In development, allow if no secret is set (for testing)
  if (!cronSecret && process.env.NODE_ENV === "development") {
    return true;
  }

  return false;
}


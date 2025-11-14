import { NextRequest, NextResponse } from "next/server";
import { verifyCronRequest } from "../utils";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "https://zaned-backennd.onrender.com";

/**
 * Save Inside Day Results Cron Job
 * Runs daily at 8:35 PM UTC (4:35 PM ET) - after market close
 * Saves inside day symbols to database
 */
export async function GET(request: NextRequest) {
  // Verify cron request to prevent unauthorized access
  if (!verifyCronRequest(request)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  try {
    const response = await fetch(`${API_BASE_URL}/api/admin/screener/save-inside-day`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(
        { error: "Backend request failed", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json({
      success: true,
      message: "Inside day results save triggered",
      data,
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    console.error("Save inside day cron error:", error);
    return NextResponse.json(
      { error: "Internal server error", message: error instanceof Error ? error.message : "Unknown error" },
      { status: 500 }
    );
  }
}


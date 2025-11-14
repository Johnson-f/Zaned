import { NextRequest, NextResponse } from "next/server";
import { verifyCronRequest } from "../utils";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "https://zaned-backennd.onrender.com";

/**
 * Historical Data Ingestion Cron Job
 * Runs every 4 hours (00:00, 04:00, 08:00, 12:00, 16:00, 20:00 UTC)
 * Fetches screener + historicals for all symbols
 */
export async function GET(request: NextRequest) {
  // Verify cron request to prevent unauthorized access
  if (!verifyCronRequest(request)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  try {
    const concurrency = 8; // Default concurrency
    const response = await fetch(
      `${API_BASE_URL}/api/admin/ingest/historicals?concurrency=${concurrency}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      }
    );

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(
        { error: "Backend request failed", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json({
      success: true,
      message: "Historical data ingestion triggered",
      data,
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    console.error("Historical data ingestion cron error:", error);
    return NextResponse.json(
      { error: "Internal server error", message: error instanceof Error ? error.message : "Unknown error" },
      { status: 500 }
    );
  }
}


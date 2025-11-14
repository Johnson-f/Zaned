import { NextRequest, NextResponse } from "next/server";
import { verifyCronRequest } from "../utils";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "https://zaned-backennd.onrender.com";

/**
 * Fundamental Data Ingestion Cron Job
 * Runs daily at 3:00 AM UTC
 * Fetches fundamental data (income, balance, cashflow) for all symbols
 * Note: Changed from weekly to daily due to Vercel Hobby plan limitations
 */
export async function GET(request: NextRequest) {
  // Verify cron request to prevent unauthorized access
  if (!verifyCronRequest(request)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  try {
    const response = await fetch(`${API_BASE_URL}/api/admin/ingest/fundamental-data`, {
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
      message: "Fundamental data ingestion triggered",
      data,
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    console.error("Fundamental data ingestion cron error:", error);
    return NextResponse.json(
      { error: "Internal server error", message: error instanceof Error ? error.message : "Unknown error" },
      { status: 500 }
    );
  }
}


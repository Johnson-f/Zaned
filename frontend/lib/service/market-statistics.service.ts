/**
 * Market Statistics Service
 * Handles all API calls to the backend market statistics endpoints
 */

import { getFullUrl, apiConfig } from "../config/api";
import type {
  MarketStatisticsResponse,
  CurrentMarketStatsResponse,
  LiveMarketStatsResponse,
} from "../types/market-statistics";

/**
 * Base fetch function with error handling
 */
async function fetchApi<T>(url: string): Promise<T> {
  const response = await fetch(url);
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || "Failed to fetch market statistics");
  }
  const data = await response.json();
  return data as T;
}

/**
 * Get historical market statistics for charting
 */
export async function getMarketStatistics(
  startDate?: string,
  endDate?: string
): Promise<MarketStatisticsResponse> {
  const params = new URLSearchParams();
  if (startDate) params.append("startDate", startDate);
  if (endDate) params.append("endDate", endDate);
  const query = params.toString();
  const url = getFullUrl(`${apiConfig.endpoints.marketStatistics.base}${query ? `?${query}` : ""}`);
  return fetchApi<MarketStatisticsResponse>(url);
}

/**
 * Get current day's real-time market statistics
 */
export async function getCurrentMarketStatistics(): Promise<CurrentMarketStatsResponse> {
  return fetchApi<CurrentMarketStatsResponse>(getFullUrl(apiConfig.endpoints.marketStatistics.current));
}

/**
 * Get live market statistics for frontend polling
 * Returns advances, decliners, unchanged, total, and last_updated timestamp
 */
export async function getLiveMarketStatistics(): Promise<LiveMarketStatsResponse> {
  return fetchApi<LiveMarketStatsResponse>(getFullUrl(apiConfig.endpoints.marketStatistics.live));
}


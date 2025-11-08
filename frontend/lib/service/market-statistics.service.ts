/**
 * Market Statistics Service
 * Handles all API calls to the backend market statistics endpoints
 */

import { API_ENDPOINTS } from "../config/api";
import type {
  MarketStatisticsResponse,
  CurrentMarketStatsResponse,
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
  const url = API_ENDPOINTS.MARKET_STATISTICS.BASE(startDate, endDate);
  return fetchApi<MarketStatisticsResponse>(url);
}

/**
 * Get current day's real-time market statistics
 */
export async function getCurrentMarketStatistics(): Promise<CurrentMarketStatsResponse> {
  return fetchApi<CurrentMarketStatsResponse>(API_ENDPOINTS.MARKET_STATISTICS.CURRENT);
}


/**
 * Type definitions for Market Statistics data models
 * Matches the backend Go model structure
 */

export interface MarketStatistics {
  id: string;
  date: string; // ISO date string (YYYY-MM-DD)
  stocksUp: number;
  stocksDown: number;
  stocksUnchanged: number;
  totalStocks: number;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface CurrentMarketStats {
  up: number;
  down: number;
  unchanged: number;
  total: number;
}

export interface LiveMarketStats {
  advances: number;
  decliners: number;
  unchanged: number;
  total: number;
  last_updated: string; // ISO 8601 timestamp
}

export interface MarketStatisticsResponse {
  success: true;
  data: MarketStatistics[];
}

export interface CurrentMarketStatsResponse {
  success: true;
  data: CurrentMarketStats;
}

export interface LiveMarketStatsResponse {
  success: true;
  data: LiveMarketStats;
}


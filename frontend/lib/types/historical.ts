/**
 * Type definitions for Historical data models
 * Matches the backend Go model structure
 */

export interface Historical {
  id: string;
  symbol: string;
  epoch: number;
  range: string;
  interval: string;
  open: number;
  high: number;
  low: number;
  close: number;
  adjClose?: number | null;
  volume: number;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface HistoricalFilterOptions {
  symbol?: string;
  minEpoch?: number;
  maxEpoch?: number;
  range?: string;
  interval?: string;
  minOpen?: number;
  maxOpen?: number;
  minHigh?: number;
  maxHigh?: number;
  minLow?: number;
  maxLow?: number;
  minClose?: number;
  maxClose?: number;
  minVolume?: number;
  maxVolume?: number;
}

export interface HistoricalSortOptions {
  field: string; // "symbol", "epoch", "open", "high", "low", "close", "volume", "created_at"
  direction: "asc" | "desc";
}

export interface HistoricalPaginationOptions {
  page: number; // 1-indexed page number
  limit: number; // Number of records per page
}

export interface HistoricalQueryResult {
  data: Historical[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface VolumeMetricsResult {
  symbol: string;
  highest_volume_in_year: number;
  highest_volume_in_quarter: number;
  highest_volume_ever: number;
}

export interface ApiError {
  success: false;
  error: string;
  message: string;
}

export interface ApiSuccess<T> {
  success: true;
  data: T;
}

export type ApiResponse<T> = ApiSuccess<T> | ApiError;


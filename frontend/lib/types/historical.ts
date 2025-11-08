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

// Generic symbols response used by screening list endpoints
export interface SymbolsResponse<Params = Record<string, unknown>> {
  symbols: string[];
  count: number;
  params?: Params;
}

// Single-symbol metric responses
export interface AdrPercentResponse<Params = Record<string, unknown>> {
  symbol: string;
  adr_percent: number;
  params: Params;
}

export interface AtrPercentResponse<Params = Record<string, unknown>> {
  symbol: string;
  atr_percent: number;
  params: Params;
}

export interface AvgVolumeDollarsResponse<Params = Record<string, unknown>> {
  symbol: string;
  avg_vol_dollars_m: number; // in millions
  params: Params;
}

export interface AvgVolumePercentResponse<Params = Record<string, unknown>> {
  symbol: string;
  vol_percent: number; // current volume as % of average
  params: Params;
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

// Screener Results types
export type ScreenerResultType = "inside_day" | "high_volume_quarter" | "high_volume_year" | "high_volume_ever";
export type ScreenerResultPeriod = "7d" | "30d" | "90d" | "ytd" | "all";

export interface ScreenerResultsResponse {
  symbols: string[];
  count: number;
  type: string;
  period: string;
}


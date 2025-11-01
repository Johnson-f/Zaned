/**
 * Type definitions for Screener data models
 * Matches the backend Go model structure
 */

export interface Screener {
  id: string;
  symbol: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  logo?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface FilterOptions {
  minPrice?: number;
  maxPrice?: number;
  minVolume?: number;
  maxVolume?: number;
  minOpen?: number;
  maxOpen?: number;
  minHigh?: number;
  maxHigh?: number;
  minLow?: number;
  maxLow?: number;
  minClose?: number;
  maxClose?: number;
}

export interface SortOptions {
  field: string; // "symbol", "open", "high", "low", "close", "volume", "created_at"
  direction: "asc" | "desc";
}

export interface PaginationOptions {
  page: number; // 1-indexed page number
  limit: number; // Number of records per page
}

export interface QueryResult {
  data: Screener[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
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


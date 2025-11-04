/**
 * Type definitions for Watchlist data models
 * Matches the backend Go model structure
 */

export interface Watchlist {
  id: string;
  user_id: string;
  name: string;
  items?: WatchlistItem[];
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface WatchlistItem {
  id: string;
  watchlist_id: string;
  symbol?: string;
  name: string;
  price?: number | null;
  afterHoursPrice?: number | null;
  change?: number | null;
  percentChange?: string;
  logo?: string;
  starred: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
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


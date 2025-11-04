/**
 * Type definitions for Company Info data models
 * Matches the backend Go model structure
 */

export interface CompanyInfo {
  symbol: string;
  name: string;
  price?: string;
  afterHoursPrice?: string;
  change?: string;
  percentChange?: string;
  open?: string;
  high?: string;
  low?: string;
  yearHigh?: string;
  yearLow?: string;
  volume?: number;
  avgVolume?: number;
  marketCap?: string;
  beta?: string;
  pe?: string;
  earningsDate?: string;
  sector?: string;
  industry?: string;
  about?: string;
  employees?: string;
  fiveDaysReturn?: string;
  oneMonthReturn?: string;
  threeMonthsReturn?: string;
  ytdReturn?: string;
  oneYearReturn?: string;
  threeYearsReturn?: string;
  fiveYearsReturn?: string;
  logo?: string;
  createdAt?: string;
  updatedAt?: string;
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


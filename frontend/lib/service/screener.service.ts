/**
 * Screener Service
 * Handles all API calls to the backend screener endpoints
 */

import { API_ENDPOINTS, createAuthHeaders } from "../config/api";
import type {
  ApiResponse,
  FilterOptions,
  PaginationOptions,
  QueryResult,
  Screener,
  SortOptions,
} from "../types/screener";

/**
 * Base fetch function with error handling
 */
async function fetchApi<T>(
  url: string,
  options?: RequestInit
): Promise<ApiResponse<T>> {
  try {
    const headers = await createAuthHeaders();
    const response = await fetch(url, {
      ...options,
      headers: {
        ...headers,
        ...options?.headers,
      },
    });

    const data = await response.json();

    if (!response.ok) {
      return {
        success: false,
        error: data.error || "Unknown error",
        message: data.message || "Request failed",
      };
    }

    return data as ApiResponse<T>;
  } catch (error) {
    return {
      success: false,
      error: "Network Error",
      message:
        error instanceof Error ? error.message : "Failed to fetch data",
    };
  }
}

/**
 * Get all screeners
 */
export async function getAllScreeners(): Promise<ApiResponse<Screener[]>> {
  return fetchApi<Screener[]>(API_ENDPOINTS.SCREENER.BASE);
}

/**
 * Get screener by ID
 */
export async function getScreenerById(
  id: string
): Promise<ApiResponse<Screener>> {
  return fetchApi<Screener>(API_ENDPOINTS.SCREENER.BY_ID(id));
}

/**
 * Get screener by symbol
 */
export async function getScreenerBySymbol(
  symbol: string
): Promise<ApiResponse<Screener>> {
  return fetchApi<Screener>(API_ENDPOINTS.SCREENER.BY_SYMBOL(symbol));
}

/**
 * Get screeners by multiple symbols
 */
export async function getScreenersBySymbols(
  symbols: string[]
): Promise<ApiResponse<Screener[]>> {
  return fetchApi<Screener[]>(API_ENDPOINTS.SCREENER.SYMBOLS, {
    method: "POST",
    body: JSON.stringify({ symbols }),
  });
}

/**
 * Get screeners with filters, sorting, and pagination
 */
export async function getScreenersWithFilters(
  filters?: FilterOptions,
  sort?: SortOptions,
  pagination?: PaginationOptions
): Promise<ApiResponse<QueryResult>> {
  const params = new URLSearchParams();

  // Add filter parameters
  if (filters) {
    if (filters.minPrice !== undefined)
      params.append("min_price", filters.minPrice.toString());
    if (filters.maxPrice !== undefined)
      params.append("max_price", filters.maxPrice.toString());
    if (filters.minVolume !== undefined)
      params.append("min_volume", filters.minVolume.toString());
    if (filters.maxVolume !== undefined)
      params.append("max_volume", filters.maxVolume.toString());
    if (filters.minOpen !== undefined)
      params.append("min_open", filters.minOpen.toString());
    if (filters.maxOpen !== undefined)
      params.append("max_open", filters.maxOpen.toString());
    if (filters.minHigh !== undefined)
      params.append("min_high", filters.minHigh.toString());
    if (filters.maxHigh !== undefined)
      params.append("max_high", filters.maxHigh.toString());
    if (filters.minLow !== undefined)
      params.append("min_low", filters.minLow.toString());
    if (filters.maxLow !== undefined)
      params.append("max_low", filters.maxLow.toString());
    if (filters.minClose !== undefined)
      params.append("min_close", filters.minClose.toString());
    if (filters.maxClose !== undefined)
      params.append("max_close", filters.maxClose.toString());
  }

  // Add sort parameters
  if (sort) {
    params.append("sort_field", sort.field);
    params.append("sort_direction", sort.direction);
  }

  // Add pagination parameters
  if (pagination) {
    params.append("page", pagination.page.toString());
    params.append("limit", pagination.limit.toString());
  }

  const url = `${API_ENDPOINTS.SCREENER.FILTER}?${params.toString()}`;
  return fetchApi<QueryResult>(url);
}

/**
 * Search screeners by symbol
 */
export async function searchScreenersBySymbol(
  searchTerm: string,
  limit: number = 10
): Promise<ApiResponse<Screener[]>> {
  const params = new URLSearchParams();
  params.append("q", searchTerm);
  params.append("limit", limit.toString());

  const url = `${API_ENDPOINTS.SCREENER.SEARCH}?${params.toString()}`;
  return fetchApi<Screener[]>(url);
}

/**
 * Get screeners by price range
 */
export async function getScreenersByPriceRange(
  minPrice: number,
  maxPrice: number
): Promise<ApiResponse<Screener[]>> {
  const params = new URLSearchParams();
  params.append("min", minPrice.toString());
  params.append("max", maxPrice.toString());

  const url = `${API_ENDPOINTS.SCREENER.PRICE_RANGE}?${params.toString()}`;
  return fetchApi<Screener[]>(url);
}

/**
 * Get screeners by volume range
 */
export async function getScreenersByVolumeRange(
  minVolume: number,
  maxVolume: number
): Promise<ApiResponse<Screener[]>> {
  const params = new URLSearchParams();
  params.append("min", minVolume.toString());
  params.append("max", maxVolume.toString());

  const url = `${API_ENDPOINTS.SCREENER.VOLUME_RANGE}?${params.toString()}`;
  return fetchApi<Screener[]>(url);
}

/**
 * Get top gainers
 */
export async function getTopGainers(
  limit: number = 10
): Promise<ApiResponse<Screener[]>> {
  const params = new URLSearchParams();
  params.append("limit", limit.toString());

  const url = `${API_ENDPOINTS.SCREENER.TOP_GAINERS}?${params.toString()}`;
  return fetchApi<Screener[]>(url);
}

/**
 * Get most active stocks
 */
export async function getMostActive(
  limit: number = 10
): Promise<ApiResponse<Screener[]>> {
  const params = new URLSearchParams();
  params.append("limit", limit.toString());

  const url = `${API_ENDPOINTS.SCREENER.MOST_ACTIVE}?${params.toString()}`;
  return fetchApi<Screener[]>(url);
}

/**
 * Get total count of screeners
 */
export async function getScreenerCount(): Promise<ApiResponse<{ count: number }>> {
  return fetchApi<{ count: number }>(API_ENDPOINTS.SCREENER.COUNT);
}


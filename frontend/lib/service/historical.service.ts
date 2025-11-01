/**
 * Historical Service
 * Handles all API calls to the backend historical endpoints
 */

import { API_ENDPOINTS, createAuthHeaders } from "../config/api";
import type {
  ApiResponse,
  Historical,
  HistoricalFilterOptions,
  HistoricalPaginationOptions,
  HistoricalQueryResult,
  HistoricalSortOptions,
  VolumeMetricsResult,
} from "../types/historical";

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
 * Get all historical records
 */
export async function getAllHistorical(): Promise<ApiResponse<Historical[]>> {
  return fetchApi<Historical[]>(API_ENDPOINTS.HISTORICAL.BASE);
}

/**
 * Get historical record by ID
 */
export async function getHistoricalById(
  id: string
): Promise<ApiResponse<Historical>> {
  return fetchApi<Historical>(API_ENDPOINTS.HISTORICAL.BY_ID(id));
}

/**
 * Get historical records by symbol
 */
export async function getHistoricalBySymbol(
  symbol: string
): Promise<ApiResponse<Historical[]>> {
  return fetchApi<Historical[]>(API_ENDPOINTS.HISTORICAL.BY_SYMBOL(symbol));
}

/**
 * Get historical records by symbol, range, and interval
 */
export async function getHistoricalBySymbolAndParams(
  symbol: string,
  range: string,
  interval: string
): Promise<ApiResponse<Historical[]>> {
  return fetchApi<Historical[]>(
    API_ENDPOINTS.HISTORICAL.BY_SYMBOL_AND_PARAMS(symbol, range, interval)
  );
}

/**
 * Get historical records with filters, sorting, and pagination
 */
export async function getHistoricalWithFilters(
  filters?: HistoricalFilterOptions,
  sort?: HistoricalSortOptions,
  pagination?: HistoricalPaginationOptions
): Promise<ApiResponse<HistoricalQueryResult>> {
  const params = new URLSearchParams();

  // Add filter parameters
  if (filters) {
    if (filters.symbol) params.append("symbol", filters.symbol);
    if (filters.minEpoch !== undefined)
      params.append("min_epoch", filters.minEpoch.toString());
    if (filters.maxEpoch !== undefined)
      params.append("max_epoch", filters.maxEpoch.toString());
    if (filters.range) params.append("range", filters.range);
    if (filters.interval) params.append("interval", filters.interval);
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
    if (filters.minVolume !== undefined)
      params.append("min_volume", filters.minVolume.toString());
    if (filters.maxVolume !== undefined)
      params.append("max_volume", filters.maxVolume.toString());
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

  const url = `${API_ENDPOINTS.HISTORICAL.FILTER}?${params.toString()}`;
  return fetchApi<HistoricalQueryResult>(url);
}

/**
 * Get total count of historical records
 */
export async function getHistoricalCount(): Promise<
  ApiResponse<{ count: number }>
> {
  return fetchApi<{ count: number }>(API_ENDPOINTS.HISTORICAL.COUNT);
}

/**
 * Get count of historical records by symbol
 */
export async function getHistoricalCountBySymbol(
  symbol: string
): Promise<ApiResponse<{ symbol: string; count: number }>> {
  return fetchApi<{ symbol: string; count: number }>(
    API_ENDPOINTS.HISTORICAL.COUNT_BY_SYMBOL(symbol)
  );
}

/**
 * Get stocks volume metrics
 */
export async function getStocksVolumeMetrics(): Promise<
  ApiResponse<VolumeMetricsResult[]>
> {
  return fetchApi<VolumeMetricsResult[]>(
    API_ENDPOINTS.HISTORICAL.VOLUME_METRICS
  );
}

/**
 * Create a historical record
 */
export async function createHistorical(
  historical: Historical
): Promise<ApiResponse<Historical>> {
  return fetchApi<Historical>(API_ENDPOINTS.HISTORICAL.BASE, {
    method: "POST",
    body: JSON.stringify(historical),
  });
}

/**
 * Create multiple historical records in batch
 */
export async function createHistoricalBatch(
  historical: Historical[]
): Promise<ApiResponse<{ message: string; count: number }>> {
  return fetchApi<{ message: string; count: number }>(
    API_ENDPOINTS.HISTORICAL.BATCH,
    {
      method: "POST",
      body: JSON.stringify(historical),
    }
  );
}

/**
 * Upsert a historical record
 */
export async function upsertHistorical(
  historical: Historical
): Promise<ApiResponse<Historical>> {
  return fetchApi<Historical>(API_ENDPOINTS.HISTORICAL.BASE, {
    method: "PUT",
    body: JSON.stringify(historical),
  });
}

/**
 * Upsert multiple historical records in batch
 */
export async function upsertHistoricalBatch(
  historical: Historical[]
): Promise<ApiResponse<{ message: string; count: number }>> {
  return fetchApi<{ message: string; count: number }>(
    API_ENDPOINTS.HISTORICAL.BATCH,
    {
      method: "PUT",
      body: JSON.stringify(historical),
    }
  );
}

/**
 * Update a historical record by ID
 */
export async function updateHistorical(
  id: string,
  historical: Partial<Historical>
): Promise<ApiResponse<{ message: string }>> {
  return fetchApi<{ message: string }>(
    API_ENDPOINTS.HISTORICAL.BY_ID(id),
    {
      method: "PUT",
      body: JSON.stringify(historical),
    }
  );
}

/**
 * Delete a historical record by ID
 */
export async function deleteHistorical(
  id: string
): Promise<ApiResponse<{ message: string }>> {
  return fetchApi<{ message: string }>(
    API_ENDPOINTS.HISTORICAL.BY_ID(id),
    {
      method: "DELETE",
    }
  );
}

/**
 * Delete all historical records by symbol
 */
export async function deleteHistoricalBySymbol(
  symbol: string
): Promise<ApiResponse<{ message: string }>> {
  return fetchApi<{ message: string }>(
    API_ENDPOINTS.HISTORICAL.DELETE_BY_SYMBOL(symbol),
    {
      method: "DELETE",
    }
  );
}

/**
 * Delete historical records by symbol, range, and interval
 */
export async function deleteHistoricalBySymbolAndParams(
  symbol: string,
  range: string,
  interval: string
): Promise<ApiResponse<{ message: string }>> {
  return fetchApi<{ message: string }>(
    API_ENDPOINTS.HISTORICAL.DELETE_BY_SYMBOL_AND_PARAMS(symbol, range, interval),
    {
      method: "DELETE",
    }
  );
}


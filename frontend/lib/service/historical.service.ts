/**
 * Historical Service
 * Handles all API calls to the backend historical endpoints
 */

import { API_ENDPOINTS, createAuthHeaders } from "../config/api";
import type {
  ApiResponse,
  Historical,
  SymbolsResponse,
  AdrPercentResponse,
  AtrPercentResponse,
  AvgVolumeDollarsResponse,
  AvgVolumePercentResponse,
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
 * Get historical records by symbol, range, and interval
 */
export async function getHistoricalBySymbol(
  symbol: string,
  range: string,
  interval: string
): Promise<ApiResponse<Historical[]>> {
  return fetchApi<Historical[]>(
    API_ENDPOINTS.HISTORICAL.BY_SYMBOL(symbol, range, interval)
  );
}

/**
 * Get screener results with time period filtering
 */
export async function getScreenerResults(
  type: string,
  period?: string
): Promise<ApiResponse<{ symbols: string[]; count: number; type: string; period: string }>> {
  const url = API_ENDPOINTS.SCREENER_RESULTS.BASE(type, period);
  return fetchApi<{ symbols: string[]; count: number; type: string; period: string }>(url);
}

export async function getAdrScreen(params: {
  range?: string;
  interval?: string;
  lookback?: number;
  minAdr?: number;
  maxAdr?: number;
}): Promise<ApiResponse<SymbolsResponse>> {
  const url = API_ENDPOINTS.SCREENING.ADR_SCREEN(
    params.range,
    params.interval,
    params.lookback,
    params.minAdr,
    params.maxAdr
  );
  return fetchApi<SymbolsResponse>(url);
}

export async function getAtrScreen(params: {
  range?: string;
  interval?: string;
  lookback?: number;
  minAtr?: number;
  maxAtr?: number;
}): Promise<ApiResponse<SymbolsResponse>> {
  const url = API_ENDPOINTS.SCREENING.ATR_SCREEN(
    params.range,
    params.interval,
    params.lookback,
    params.minAtr,
    params.maxAtr
  );
  return fetchApi<SymbolsResponse>(url);
}

export async function getAdrForSymbol(params: {
  symbol: string;
  range?: string;
  interval?: string;
  lookback?: number;
}): Promise<ApiResponse<AdrPercentResponse>> {
  const url = API_ENDPOINTS.SCREENING.ADR(
    params.symbol,
    params.range,
    params.interval,
    params.lookback
  );
  return fetchApi<AdrPercentResponse>(url);
}

export async function getAtrForSymbol(params: {
  symbol: string;
  range?: string;
  interval?: string;
  lookback?: number;
}): Promise<ApiResponse<AtrPercentResponse>> {
  const url = API_ENDPOINTS.SCREENING.ATR(
    params.symbol,
    params.range,
    params.interval,
    params.lookback
  );
  return fetchApi<AtrPercentResponse>(url);
}

export async function getAvgVolumeDollarsScreen(params: {
  range?: string;
  interval?: string;
  lookback?: number;
  minVolDollarsM?: number;
  maxVolDollarsM?: number;
}): Promise<ApiResponse<SymbolsResponse>> {
  const url = API_ENDPOINTS.SCREENING.AVG_VOLUME_DOLLARS_SCREEN(
    params.range,
    params.interval,
    params.lookback,
    params.minVolDollarsM,
    params.maxVolDollarsM
  );
  return fetchApi<SymbolsResponse>(url);
}

export async function getAvgVolumePercentScreen(params: {
  range?: string;
  interval?: string;
  lookback?: number;
  minVolPercent?: number;
  maxVolPercent?: number;
}): Promise<ApiResponse<SymbolsResponse>> {
  const url = API_ENDPOINTS.SCREENING.AVG_VOLUME_PERCENT_SCREEN(
    params.range,
    params.interval,
    params.lookback,
    params.minVolPercent,
    params.maxVolPercent
  );
  return fetchApi<SymbolsResponse>(url);
}

export async function getAvgVolumeDollarsForSymbol(params: {
  symbol: string;
  range?: string;
  interval?: string;
  lookback?: number;
}): Promise<ApiResponse<AvgVolumeDollarsResponse>> {
  const url = API_ENDPOINTS.SCREENING.AVG_VOLUME_DOLLARS(
    params.symbol,
    params.range,
    params.interval,
    params.lookback
  );
  return fetchApi<AvgVolumeDollarsResponse>(url);
}

export async function getAvgVolumePercentForSymbol(params: {
  symbol: string;
  range?: string;
  interval?: string;
  lookback?: number;
}): Promise<ApiResponse<AvgVolumePercentResponse>> {
  const url = API_ENDPOINTS.SCREENING.AVG_VOLUME_PERCENT(
    params.symbol,
    params.range,
    params.interval,
    params.lookback
  );
  return fetchApi<AvgVolumePercentResponse>(url);
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

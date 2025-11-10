/**
 * Historical Service
 * Handles all API calls to the backend historical endpoints
 */

import { getFullUrl, apiConfig, createAuthHeaders } from "../config/api";
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
  return fetchApi<Historical[]>(getFullUrl(apiConfig.endpoints.historical.base));
}

/**
 * Get historical record by ID
 */
export async function getHistoricalById(
  id: string
): Promise<ApiResponse<Historical>> {
  return fetchApi<Historical>(getFullUrl(apiConfig.endpoints.historical.byId(id)));
}

/**
 * Get historical records by symbol, range, and interval
 */
export async function getHistoricalBySymbol(
  symbol: string,
  range: string,
  interval: string
): Promise<ApiResponse<Historical[]>> {
  const params = new URLSearchParams();
  params.append("symbol", symbol);
  params.append("range", range);
  params.append("interval", interval);
  return fetchApi<Historical[]>(
    getFullUrl(`${apiConfig.endpoints.historical.bySymbol}?${params.toString()}`)
  );
}

/**
 * Get screener results with time period filtering
 */
export async function getScreenerResults(
  type: string,
  period?: string
): Promise<ApiResponse<{ symbols: string[]; count: number; type: string; period: string }>> {
  const params = new URLSearchParams();
  params.append("type", type);
  if (period) params.append("period", period);
  const url = getFullUrl(`${apiConfig.endpoints.screenerResults.base}?${params.toString()}`);
  return fetchApi<{ symbols: string[]; count: number; type: string; period: string }>(url);
}

export async function getAdrScreen(params: {
  range?: string;
  interval?: string;
  lookback?: number;
  minAdr?: number;
  maxAdr?: number;
}): Promise<ApiResponse<SymbolsResponse>> {
  const queryParams = new URLSearchParams();
  if (params.range) queryParams.append("range", params.range);
  if (params.interval) queryParams.append("interval", params.interval);
  if (params.lookback) queryParams.append("lookback", params.lookback.toString());
  if (params.minAdr !== undefined) queryParams.append("min_adr", params.minAdr.toString());
  if (params.maxAdr !== undefined) queryParams.append("max_adr", params.maxAdr.toString());
  const query = queryParams.toString();
  const url = getFullUrl(`${apiConfig.endpoints.screening.adrScreen}${query ? `?${query}` : ""}`);
  return fetchApi<SymbolsResponse>(url);
}

export async function getAtrScreen(params: {
  range?: string;
  interval?: string;
  lookback?: number;
  minAtr?: number;
  maxAtr?: number;
}): Promise<ApiResponse<SymbolsResponse>> {
  const queryParams = new URLSearchParams();
  if (params.range) queryParams.append("range", params.range);
  if (params.interval) queryParams.append("interval", params.interval);
  if (params.lookback) queryParams.append("lookback", params.lookback.toString());
  if (params.minAtr !== undefined) queryParams.append("min_atr", params.minAtr.toString());
  if (params.maxAtr !== undefined) queryParams.append("max_atr", params.maxAtr.toString());
  const query = queryParams.toString();
  const url = getFullUrl(`${apiConfig.endpoints.screening.atrScreen}${query ? `?${query}` : ""}`);
  return fetchApi<SymbolsResponse>(url);
}

export async function getAdrForSymbol(params: {
  symbol: string;
  range?: string;
  interval?: string;
  lookback?: number;
}): Promise<ApiResponse<AdrPercentResponse>> {
  const queryParams = new URLSearchParams();
  queryParams.append("symbol", params.symbol);
  if (params.range) queryParams.append("range", params.range);
  if (params.interval) queryParams.append("interval", params.interval);
  if (params.lookback) queryParams.append("lookback", params.lookback.toString());
  const url = getFullUrl(`${apiConfig.endpoints.screening.adr}?${queryParams.toString()}`);
  return fetchApi<AdrPercentResponse>(url);
}

export async function getAtrForSymbol(params: {
  symbol: string;
  range?: string;
  interval?: string;
  lookback?: number;
}): Promise<ApiResponse<AtrPercentResponse>> {
  const queryParams = new URLSearchParams();
  queryParams.append("symbol", params.symbol);
  if (params.range) queryParams.append("range", params.range);
  if (params.interval) queryParams.append("interval", params.interval);
  if (params.lookback) queryParams.append("lookback", params.lookback.toString());
  const url = getFullUrl(`${apiConfig.endpoints.screening.atr}?${queryParams.toString()}`);
  return fetchApi<AtrPercentResponse>(url);
}

export async function getAvgVolumeDollarsScreen(params: {
  range?: string;
  interval?: string;
  lookback?: number;
  minVolDollarsM?: number;
  maxVolDollarsM?: number;
}): Promise<ApiResponse<SymbolsResponse>> {
  const queryParams = new URLSearchParams();
  if (params.range) queryParams.append("range", params.range);
  if (params.interval) queryParams.append("interval", params.interval);
  if (params.lookback) queryParams.append("lookback", params.lookback.toString());
  if (params.minVolDollarsM !== undefined) queryParams.append("min_vol_dollars_m", params.minVolDollarsM.toString());
  if (params.maxVolDollarsM !== undefined) queryParams.append("max_vol_dollars_m", params.maxVolDollarsM.toString());
  const query = queryParams.toString();
  const url = getFullUrl(`${apiConfig.endpoints.screening.avgVolumeDollarsScreen}${query ? `?${query}` : ""}`);
  return fetchApi<SymbolsResponse>(url);
}

export async function getAvgVolumePercentScreen(params: {
  range?: string;
  interval?: string;
  lookback?: number;
  minVolPercent?: number;
  maxVolPercent?: number;
}): Promise<ApiResponse<SymbolsResponse>> {
  const queryParams = new URLSearchParams();
  if (params.range) queryParams.append("range", params.range);
  if (params.interval) queryParams.append("interval", params.interval);
  if (params.lookback) queryParams.append("lookback", params.lookback.toString());
  if (params.minVolPercent !== undefined) queryParams.append("min_vol_percent", params.minVolPercent.toString());
  if (params.maxVolPercent !== undefined) queryParams.append("max_vol_percent", params.maxVolPercent.toString());
  const query = queryParams.toString();
  const url = getFullUrl(`${apiConfig.endpoints.screening.avgVolumePercentScreen}${query ? `?${query}` : ""}`);
  return fetchApi<SymbolsResponse>(url);
}

export async function getAvgVolumeDollarsForSymbol(params: {
  symbol: string;
  range?: string;
  interval?: string;
  lookback?: number;
}): Promise<ApiResponse<AvgVolumeDollarsResponse>> {
  const queryParams = new URLSearchParams();
  queryParams.append("symbol", params.symbol);
  if (params.range) queryParams.append("range", params.range);
  if (params.interval) queryParams.append("interval", params.interval);
  if (params.lookback) queryParams.append("lookback", params.lookback.toString());
  const url = getFullUrl(`${apiConfig.endpoints.screening.avgVolumeDollars}?${queryParams.toString()}`);
  return fetchApi<AvgVolumeDollarsResponse>(url);
}

export async function getAvgVolumePercentForSymbol(params: {
  symbol: string;
  range?: string;
  interval?: string;
  lookback?: number;
}): Promise<ApiResponse<AvgVolumePercentResponse>> {
  const queryParams = new URLSearchParams();
  queryParams.append("symbol", params.symbol);
  if (params.range) queryParams.append("range", params.range);
  if (params.interval) queryParams.append("interval", params.interval);
  if (params.lookback) queryParams.append("lookback", params.lookback.toString());
  const url = getFullUrl(`${apiConfig.endpoints.screening.avgVolumePercent}?${queryParams.toString()}`);
  return fetchApi<AvgVolumePercentResponse>(url);
}

/**
 * Create a historical record
 */
export async function createHistorical(
  historical: Historical
): Promise<ApiResponse<Historical>> {
  return fetchApi<Historical>(getFullUrl(apiConfig.endpoints.historical.base), {
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
    getFullUrl(apiConfig.endpoints.historical.batch),
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
  return fetchApi<Historical>(getFullUrl(apiConfig.endpoints.historical.base), {
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
    getFullUrl(apiConfig.endpoints.historical.batch),
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
    getFullUrl(apiConfig.endpoints.historical.byId(id)),
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
    getFullUrl(apiConfig.endpoints.historical.byId(id)),
    {
      method: "DELETE",
    }
  );
}

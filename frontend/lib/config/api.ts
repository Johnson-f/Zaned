/**
 * API Configuration
 * Centralized configuration for backend API connection
 */

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const API_ENDPOINTS = {
  // Public routes
  HEALTH: `${API_BASE_URL}/api/health`,

  // Public screening endpoints (no authentication required)
  SCREENING: {
    // Inside day detection
    INSIDE_DAY: `${API_BASE_URL}/api/inside-day`,

    // Highest volume detection
    HIGH_VOLUME_QUARTER: `${API_BASE_URL}/api/high-volume-quarter`,
    HIGH_VOLUME_YEAR: `${API_BASE_URL}/api/high-volume-year`,
    HIGH_VOLUME_EVER: `${API_BASE_URL}/api/high-volume-ever`,

    // ADR (Average Daily Range) screening
    ADR_SCREEN: (range?: string, interval?: string, lookback?: number, minAdr?: number, maxAdr?: number) => {
      const params = new URLSearchParams();
      if (range) params.append("range", range);
      if (interval) params.append("interval", interval);
      if (lookback) params.append("lookback", lookback.toString());
      if (minAdr !== undefined) params.append("min_adr", minAdr.toString());
      if (maxAdr !== undefined) params.append("max_adr", maxAdr.toString());
      const query = params.toString();
      return `${API_BASE_URL}/api/adr-screen${query ? `?${query}` : ""}`;
    },
    ADR: (symbol: string, range?: string, interval?: string, lookback?: number) => {
      const params = new URLSearchParams();
      params.append("symbol", symbol);
      if (range) params.append("range", range);
      if (interval) params.append("interval", interval);
      if (lookback) params.append("lookback", lookback.toString());
      return `${API_BASE_URL}/api/adr?${params.toString()}`;
    },

    // ATR (Average True Range) screening
    ATR_SCREEN: (range?: string, interval?: string, lookback?: number, minAtr?: number, maxAtr?: number) => {
      const params = new URLSearchParams();
      if (range) params.append("range", range);
      if (interval) params.append("interval", interval);
      if (lookback) params.append("lookback", lookback.toString());
      if (minAtr !== undefined) params.append("min_atr", minAtr.toString());
      if (maxAtr !== undefined) params.append("max_atr", maxAtr.toString());
      const query = params.toString();
      return `${API_BASE_URL}/api/atr-screen${query ? `?${query}` : ""}`;
    },
    ATR: (symbol: string, range?: string, interval?: string, lookback?: number) => {
      const params = new URLSearchParams();
      params.append("symbol", symbol);
      if (range) params.append("range", range);
      if (interval) params.append("interval", interval);
      if (lookback) params.append("lookback", lookback.toString());
      return `${API_BASE_URL}/api/atr?${params.toString()}`;
    },

    // Average volume in dollars screening
    AVG_VOLUME_DOLLARS_SCREEN: (range?: string, interval?: string, lookback?: number, minVolDollarsM?: number, maxVolDollarsM?: number) => {
      const params = new URLSearchParams();
      if (range) params.append("range", range);
      if (interval) params.append("interval", interval);
      if (lookback) params.append("lookback", lookback.toString());
      if (minVolDollarsM !== undefined) params.append("min_vol_dollars_m", minVolDollarsM.toString());
      if (maxVolDollarsM !== undefined) params.append("max_vol_dollars_m", maxVolDollarsM.toString());
      const query = params.toString();
      return `${API_BASE_URL}/api/avg-volume-dollars-screen${query ? `?${query}` : ""}`;
    },
    AVG_VOLUME_DOLLARS: (symbol: string, range?: string, interval?: string, lookback?: number) => {
      const params = new URLSearchParams();
      params.append("symbol", symbol);
      if (range) params.append("range", range);
      if (interval) params.append("interval", interval);
      if (lookback) params.append("lookback", lookback.toString());
      return `${API_BASE_URL}/api/avg-volume-dollars?${params.toString()}`;
    },

    // Average volume in percent screening
    AVG_VOLUME_PERCENT_SCREEN: (range?: string, interval?: string, lookback?: number, minVolPercent?: number, maxVolPercent?: number) => {
      const params = new URLSearchParams();
      if (range) params.append("range", range);
      if (interval) params.append("interval", interval);
      if (lookback) params.append("lookback", lookback.toString());
      if (minVolPercent !== undefined) params.append("min_vol_percent", minVolPercent.toString());
      if (maxVolPercent !== undefined) params.append("max_vol_percent", maxVolPercent.toString());
      const query = params.toString();
      return `${API_BASE_URL}/api/avg-volume-percent-screen${query ? `?${query}` : ""}`;
    },
    AVG_VOLUME_PERCENT: (symbol: string, range?: string, interval?: string, lookback?: number) => {
      const params = new URLSearchParams();
      params.append("symbol", symbol);
      if (range) params.append("range", range);
      if (interval) params.append("interval", interval);
      if (lookback) params.append("lookback", lookback.toString());
      return `${API_BASE_URL}/api/avg-volume-percent?${params.toString()}`;
    },
  },

  // Admin endpoints (public, no authentication)
  ADMIN: {
    INGEST_HISTORICALS: (concurrency?: number) => {
      const params = new URLSearchParams();
      if (concurrency) params.append("concurrency", concurrency.toString());
      const query = params.toString();
      return `${API_BASE_URL}/api/admin/ingest/historicals${query ? `?${query}` : ""}`;
    },
  },

  // Protected routes (require JWT authentication)
  SCREENER: {
    BASE: `${API_BASE_URL}/api/protected/screener`,
    BY_ID: (id: string) => `${API_BASE_URL}/api/protected/screener/${id}`,
    BY_SYMBOL: (symbol: string) =>
      `${API_BASE_URL}/api/protected/screener/symbol/${symbol}`,
    FILTER: `${API_BASE_URL}/api/protected/screener/filter`,
    SEARCH: `${API_BASE_URL}/api/protected/screener/search`,
    PRICE_RANGE: `${API_BASE_URL}/api/protected/screener/price-range`,
    VOLUME_RANGE: `${API_BASE_URL}/api/protected/screener/volume-range`,
    TOP_GAINERS: `${API_BASE_URL}/api/protected/screener/top-gainers`,
    MOST_ACTIVE: `${API_BASE_URL}/api/protected/screener/most-active`,
    COUNT: `${API_BASE_URL}/api/protected/screener/count`,
    SYMBOLS: `${API_BASE_URL}/api/protected/screener/symbols`,
  },
  HISTORICAL: {
    BASE: `${API_BASE_URL}/api/protected/historical`,
    BY_ID: (id: string) => `${API_BASE_URL}/api/protected/historical/${id}`,
    BATCH: `${API_BASE_URL}/api/protected/historical/batch`,
  },
} as const;

/**
 * Get authentication token from Supabase session
 */
export async function getAuthToken(): Promise<string | null> {
  try {
    const { createClient } = await import("../supabase/client");
    const supabase = createClient();
    const {
      data: { session },
    } = await supabase.auth.getSession();
    return session?.access_token || null;
  } catch (error) {
    console.error("Error getting auth token:", error);
    return null;
  }
}

/**
 * Create headers for API requests with authentication
 */
export async function createAuthHeaders(): Promise<HeadersInit> {
  const token = await getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return headers;
}


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

  // Company Info endpoints (public, read-only)
  COMPANY_INFO: {
    BASE: `${API_BASE_URL}/api/company-info`,
    BY_SYMBOL: (symbol: string) => `${API_BASE_URL}/api/company-info/${symbol}`,
    BY_SYMBOLS: `${API_BASE_URL}/api/company-info/symbols`,
    SEARCH: (query: string) => `${API_BASE_URL}/api/company-info/search?q=${encodeURIComponent(query)}`,
    BY_SECTOR: (sector: string) => `${API_BASE_URL}/api/company-info/sector/${encodeURIComponent(sector)}`,
    BY_INDUSTRY: (industry: string) => `${API_BASE_URL}/api/company-info/industry/${encodeURIComponent(industry)}`,
  },

  // Fundamental Data endpoints (public, read-only)
  FUNDAMENTAL_DATA: {
    BASE: `${API_BASE_URL}/api/fundamental-data`,
    BY_SYMBOL: (symbol: string) => `${API_BASE_URL}/api/fundamental-data/symbol/${encodeURIComponent(symbol)}`,
    BY_SYMBOL_AND_TYPE: (symbol: string, statementType: string) =>
      `${API_BASE_URL}/api/fundamental-data/symbol/${encodeURIComponent(symbol)}/type/${encodeURIComponent(statementType)}`,
    BY_SYMBOL_TYPE_AND_FREQUENCY: (symbol: string, statementType: string, frequency: string) =>
      `${API_BASE_URL}/api/fundamental-data/symbol/${encodeURIComponent(symbol)}/type/${encodeURIComponent(statementType)}/frequency/${encodeURIComponent(frequency)}`,
    BY_STATEMENT_TYPE: (statementType: string) =>
      `${API_BASE_URL}/api/fundamental-data/type/${encodeURIComponent(statementType)}`,
    BY_FREQUENCY: (frequency: string) =>
      `${API_BASE_URL}/api/fundamental-data/frequency/${encodeURIComponent(frequency)}`,
    SEARCH: (query: string) =>
      `${API_BASE_URL}/api/fundamental-data/search?q=${encodeURIComponent(query)}`,
    METRICS: (symbol: string, statementType?: string, frequency?: string) => {
      const params = new URLSearchParams();
      params.append("symbol", symbol);
      if (statementType) params.append("statement_type", statementType);
      if (frequency) params.append("frequency", frequency);
      return `${API_BASE_URL}/api/fundamental-data/metrics?${params.toString()}`;
    },
    REVENUE_GROWTH: (statementType?: string, frequency?: string, minQoQGrowth?: number, maxQoQGrowth?: number, minYoYGrowth?: number, maxYoYGrowth?: number) => {
      const params = new URLSearchParams();
      if (statementType) params.append("statement_type", statementType);
      if (frequency) params.append("frequency", frequency);
      if (minQoQGrowth !== undefined) params.append("min_qoq_growth", minQoQGrowth.toString());
      if (maxQoQGrowth !== undefined) params.append("max_qoq_growth", maxQoQGrowth.toString());
      if (minYoYGrowth !== undefined) params.append("min_yoy_growth", minYoYGrowth.toString());
      if (maxYoYGrowth !== undefined) params.append("max_yoy_growth", maxYoYGrowth.toString());
      return `${API_BASE_URL}/api/fundamental-data/revenue-growth?${params.toString()}`;
    },
    EPS_FILTER: (statementType?: string, frequency?: string, date?: string, minEPS?: number, maxEPS?: number) => {
      const params = new URLSearchParams();
      if (statementType) params.append("statement_type", statementType);
      if (frequency) params.append("frequency", frequency);
      if (date) params.append("date", date);
      if (minEPS !== undefined) params.append("min_eps", minEPS.toString());
      if (maxEPS !== undefined) params.append("max_eps", maxEPS.toString());
      return `${API_BASE_URL}/api/fundamental-data/eps-filter?${params.toString()}`;
    },
    MARGIN_FILTER: (marginType?: string, statementType?: string, frequency?: string, date?: string, minMargin?: number, maxMargin?: number) => {
      const params = new URLSearchParams();
      if (marginType) params.append("margin_type", marginType);
      if (statementType) params.append("statement_type", statementType);
      if (frequency) params.append("frequency", frequency);
      if (date) params.append("date", date);
      if (minMargin !== undefined) params.append("min_margin", minMargin.toString());
      if (maxMargin !== undefined) params.append("max_margin", maxMargin.toString());
      return `${API_BASE_URL}/api/fundamental-data/margin-filter?${params.toString()}`;
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
    UPDATE_WATCHLIST_PRICES: `${API_BASE_URL}/api/admin/watchlist/update-prices`,
    INGEST_COMPANY_INFO: `${API_BASE_URL}/api/admin/ingest/company-info`,
    INGEST_FUNDAMENTAL_DATA: `${API_BASE_URL}/api/admin/ingest/fundamental-data`,
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
    BY_SYMBOL: (symbol: string, range: string, interval: string) => {
      const params = new URLSearchParams();
      params.append("symbol", symbol);
      params.append("range", range);
      params.append("interval", interval);
      return `${API_BASE_URL}/api/protected/historical/by-symbol?${params.toString()}`;
    },
    BATCH: `${API_BASE_URL}/api/protected/historical/batch`,
  },
  WATCHLIST: {
    // Watchlist CRUD
    BASE: `${API_BASE_URL}/api/protected/watchlist`,
    BY_ID: (id: string) => `${API_BASE_URL}/api/protected/watchlist/${id}`,
    
    // Watchlist Items
    ITEMS: (watchlistId: string) => `${API_BASE_URL}/api/protected/watchlist/${watchlistId}/items`,
    ITEM_BY_ID: (id: string) => `${API_BASE_URL}/api/protected/watchlist/item/${id}`,
    ADD_ITEM: (watchlistId: string) => `${API_BASE_URL}/api/protected/watchlist/${watchlistId}/items`,
    UPDATE_ITEM: (id: string) => `${API_BASE_URL}/api/protected/watchlist/item/${id}`,
    DELETE_ITEM: (id: string) => `${API_BASE_URL}/api/protected/watchlist/item/${id}`,
    
    // Starred items
    TOGGLE_STAR: (id: string) => `${API_BASE_URL}/api/protected/watchlist/item/${id}/star`,
    STARRED: `${API_BASE_URL}/api/protected/watchlist/starred`,
    
    // Batch operations
    BATCH_UPDATE_ITEMS: `${API_BASE_URL}/api/protected/watchlist/items/batch`,
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


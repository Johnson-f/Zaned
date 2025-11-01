/**
 * API Configuration
 * Centralized configuration for backend API connection
 */

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const API_ENDPOINTS = {
  // Public routes
  HEALTH: `${API_BASE_URL}/api/health`,

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


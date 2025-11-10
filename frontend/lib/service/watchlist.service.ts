/**
 * Watchlist Service
 * Handles all API calls to the backend watchlist endpoints
 */

import { getFullUrl, apiConfig, createAuthHeaders } from "../config/api";
import type {
  ApiResponse,
  Watchlist,
  WatchlistItem,
} from "../types/watchlist";

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
 * Get all watchlists for the authenticated user
 */
export async function getAllWatchlists(): Promise<ApiResponse<Watchlist[]>> {
  return fetchApi<Watchlist[]>(getFullUrl(apiConfig.endpoints.watchlist.base));
}

/**
 * Get watchlist by ID
 */
export async function getWatchlistById(
  id: string
): Promise<ApiResponse<Watchlist>> {
  return fetchApi<Watchlist>(getFullUrl(apiConfig.endpoints.watchlist.byId(id)));
}

/**
 * Create a new watchlist
 */
export async function createWatchlist(
  name: string
): Promise<ApiResponse<Watchlist>> {
  return fetchApi<Watchlist>(getFullUrl(apiConfig.endpoints.watchlist.base), {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

/**
 * Update a watchlist
 */
export async function updateWatchlist(
  id: string,
  name: string
): Promise<ApiResponse<Watchlist>> {
  return fetchApi<Watchlist>(getFullUrl(apiConfig.endpoints.watchlist.byId(id)), {
    method: "PUT",
    body: JSON.stringify({ name }),
  });
}

/**
 * Delete a watchlist
 */
export async function deleteWatchlist(
  id: string
): Promise<ApiResponse<void>> {
  return fetchApi<void>(getFullUrl(apiConfig.endpoints.watchlist.byId(id)), {
    method: "DELETE",
  });
}

/**
 * Get all items for a watchlist
 */
export async function getWatchlistItems(
  watchlistId: string
): Promise<ApiResponse<WatchlistItem[]>> {
  return fetchApi<WatchlistItem[]>(
    getFullUrl(apiConfig.endpoints.watchlist.items(watchlistId))
  );
}

/**
 * Add item to watchlist
 */
export async function addItemToWatchlist(
  watchlistId: string,
  item: {
    symbol?: string;
    name: string;
    price?: number;
    afterHoursPrice?: number;
    change?: number;
    percentChange?: string;
    logo?: string;
  }
): Promise<ApiResponse<WatchlistItem>> {
  return fetchApi<WatchlistItem>(
    getFullUrl(apiConfig.endpoints.watchlist.addItem(watchlistId)),
    {
      method: "POST",
      body: JSON.stringify(item),
    }
  );
}

/**
 * Update watchlist item
 */
export async function updateWatchlistItem(
  id: string,
  updates: Partial<WatchlistItem>
): Promise<ApiResponse<WatchlistItem>> {
  return fetchApi<WatchlistItem>(getFullUrl(apiConfig.endpoints.watchlist.updateItem(id)), {
    method: "PUT",
    body: JSON.stringify(updates),
  });
}

/**
 * Delete watchlist item
 */
export async function deleteWatchlistItem(
  id: string
): Promise<ApiResponse<void>> {
  return fetchApi<void>(getFullUrl(apiConfig.endpoints.watchlist.deleteItem(id)), {
    method: "DELETE",
  });
}

/**
 * Toggle star on watchlist item
 */
export async function toggleStarOnItem(
  id: string
): Promise<ApiResponse<WatchlistItem>> {
  return fetchApi<WatchlistItem>(getFullUrl(apiConfig.endpoints.watchlist.toggleStar(id)), {
    method: "POST",
  });
}


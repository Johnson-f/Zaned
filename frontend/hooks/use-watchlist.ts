/**
 * TanStack Query hooks for Watchlist API
 * Provides reactive data fetching with caching and error handling
 */

import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import type {
  Watchlist,
  WatchlistItem,
} from "../lib/types/watchlist";
import * as watchlistService from "../lib/service/watchlist.service";

/**
 * Cache configuration constants
 */
const CACHE_CONFIG = {
  // List queries (multiple items) - cache for 5 minutes, stale after 2 minutes
  LIST: {
    staleTime: 2 * 60 * 1000, // 2 minutes
    gcTime: 5 * 60 * 1000, // 5 minutes (garbage collection time)
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
  },
  // Detail queries (single item) - cache for 10 minutes, stale after 5 minutes
  DETAIL: {
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
  },
  // Real-time queries - refetch every 30 seconds
  REALTIME: {
    staleTime: 30 * 1000, // 30 seconds
    gcTime: 2 * 60 * 1000, // 2 minutes
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
  },
} as const;

/**
 * Query keys for TanStack Query
 */
export const watchlistKeys = {
  all: ["watchlists"] as const,
  lists: () => [...watchlistKeys.all, "list"] as const,
  details: () => [...watchlistKeys.all, "detail"] as const,
  detail: (id: string) => [...watchlistKeys.details(), id] as const,
  items: (watchlistId: string) => [...watchlistKeys.all, "items", watchlistId] as const,
};

/**
 * Hook to get all watchlists for the authenticated user
 */
export function useWatchlists(enabled: boolean = true) {
  return useQuery({
    queryKey: watchlistKeys.lists(),
    queryFn: async () => {
      const response = await watchlistService.getAllWatchlists();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch watchlists");
      }
      return response.data || [];
    },
    enabled,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get watchlist by ID
 */
export function useWatchlistById(id: string, enabled: boolean = true) {
  return useQuery({
    queryKey: watchlistKeys.detail(id),
    queryFn: async () => {
      const response = await watchlistService.getWatchlistById(id);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch watchlist");
      }
      return response.data;
    },
    enabled: enabled && !!id,
    ...CACHE_CONFIG.DETAIL,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get watchlist items
 */
export function useWatchlistItems(
  watchlistId: string,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: watchlistKeys.items(watchlistId),
    queryFn: async () => {
      const response = await watchlistService.getWatchlistItems(watchlistId);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch watchlist items");
      }
      return response.data || [];
    },
    enabled: enabled && !!watchlistId,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to create a new watchlist
 */
export function useCreateWatchlist() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (name: string) => {
      const response = await watchlistService.createWatchlist(name);
      if (!response.success) {
        throw new Error(response.message || "Failed to create watchlist");
      }
      return response.data;
    },
    onSuccess: () => {
      // Invalidate watchlists list to refetch
      queryClient.invalidateQueries({ queryKey: watchlistKeys.lists() });
    },
  });
}

/**
 * Hook to update a watchlist
 */
export function useUpdateWatchlist() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, name }: { id: string; name: string }) => {
      const response = await watchlistService.updateWatchlist(id, name);
      if (!response.success) {
        throw new Error(response.message || "Failed to update watchlist");
      }
      return response.data;
    },
    onSuccess: (data) => {
      // Invalidate both the list and the specific watchlist
      queryClient.invalidateQueries({ queryKey: watchlistKeys.lists() });
      queryClient.invalidateQueries({ queryKey: watchlistKeys.detail(data.id) });
    },
  });
}

/**
 * Hook to delete a watchlist
 */
export function useDeleteWatchlist() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await watchlistService.deleteWatchlist(id);
      if (!response.success) {
        throw new Error(response.message || "Failed to delete watchlist");
      }
      return id;
    },
    onSuccess: (id) => {
      // Invalidate lists and remove the specific watchlist from cache
      queryClient.invalidateQueries({ queryKey: watchlistKeys.lists() });
      queryClient.removeQueries({ queryKey: watchlistKeys.detail(id) });
    },
  });
}

/**
 * Hook to add item to watchlist
 */
export function useAddWatchlistItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      watchlistId,
      item,
    }: {
      watchlistId: string;
      item: {
        symbol?: string;
        name: string;
        price?: number;
        afterHoursPrice?: number;
        change?: number;
        percentChange?: string;
        logo?: string;
      };
    }) => {
      const response = await watchlistService.addItemToWatchlist(watchlistId, item);
      if (!response.success) {
        throw new Error(response.message || "Failed to add item to watchlist");
      }
      return response.data;
    },
    onSuccess: (_data, variables) => {
      // Invalidate watchlist items and the watchlist itself
      queryClient.invalidateQueries({
        queryKey: watchlistKeys.items(variables.watchlistId),
      });
      queryClient.invalidateQueries({
        queryKey: watchlistKeys.detail(variables.watchlistId),
      });
    },
  });
}

/**
 * Hook to update watchlist item
 */
export function useUpdateWatchlistItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      id,
      updates,
    }: {
      id: string;
      updates: Partial<WatchlistItem>;
    }) => {
      const response = await watchlistService.updateWatchlistItem(id, updates);
      if (!response.success) {
        throw new Error(response.message || "Failed to update watchlist item");
      }
      return response.data;
    },
    onSuccess: (data) => {
      // Invalidate items for the watchlist
      queryClient.invalidateQueries({
        queryKey: watchlistKeys.items(data.watchlist_id),
      });
    },
  });
}

/**
 * Hook to delete watchlist item
 */
export function useDeleteWatchlistItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await watchlistService.deleteWatchlistItem(id);
      if (!response.success) {
        throw new Error(response.message || "Failed to delete watchlist item");
      }
      return id;
    },
    onSuccess: () => {
      // Invalidate all watchlist items queries
      queryClient.invalidateQueries({ queryKey: watchlistKeys.all });
    },
  });
}

/**
 * Hook to toggle star on watchlist item
 */
export function useToggleStarOnItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await watchlistService.toggleStarOnItem(id);
      if (!response.success) {
        throw new Error(response.message || "Failed to toggle star");
      }
      return response.data;
    },
    onSuccess: (data) => {
      // Invalidate items for the watchlist
      queryClient.invalidateQueries({
        queryKey: watchlistKeys.items(data.watchlist_id),
      });
    },
  });
}


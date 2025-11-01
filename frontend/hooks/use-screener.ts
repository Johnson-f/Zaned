/**
 * TanStack Query hooks for Screener API
 * Provides reactive data fetching with caching and error handling
 */

import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import type {
  FilterOptions,
  PaginationOptions,
  Screener,
  SortOptions,
} from "../lib/types/screener";
import * as screenerService from "../lib/service/screener.service";

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
  // Real-time queries (top gainers, most active) - refetch every 30 seconds
  REALTIME: {
    staleTime: 30 * 1000, // 30 seconds
    gcTime: 2 * 60 * 1000, // 2 minutes
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
  },
  // Search queries - cache for 1 minute, stale immediately
  SEARCH: {
    staleTime: 0, // Stale immediately for search results
    gcTime: 1 * 60 * 1000, // 1 minute
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  },
  // Count query - cache for 5 minutes, stale after 3 minutes
  COUNT: {
    staleTime: 3 * 60 * 1000, // 3 minutes
    gcTime: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
  },
} as const;

/**
 * Query keys for TanStack Query
 */
export const screenerKeys = {
  all: ["screeners"] as const,
  lists: () => [...screenerKeys.all, "list"] as const,
  list: (filters?: FilterOptions, sort?: SortOptions, pagination?: PaginationOptions) =>
    [...screenerKeys.lists(), filters, sort, pagination] as const,
  details: () => [...screenerKeys.all, "detail"] as const,
  detail: (id: string) => [...screenerKeys.details(), id] as const,
  bySymbol: (symbol: string) => [...screenerKeys.all, "symbol", symbol] as const,
  bySymbols: (symbols: string[]) => [...screenerKeys.all, "symbols", symbols] as const,
  search: (term: string, limit?: number) =>
    [...screenerKeys.all, "search", term, limit] as const,
  priceRange: (min: number, max: number) =>
    [...screenerKeys.all, "price-range", min, max] as const,
  volumeRange: (min: number, max: number) =>
    [...screenerKeys.all, "volume-range", min, max] as const,
  topGainers: (limit?: number) => [...screenerKeys.all, "top-gainers", limit] as const,
  mostActive: (limit?: number) => [...screenerKeys.all, "most-active", limit] as const,
  count: () => [...screenerKeys.all, "count"] as const,
};

/**
 * Hook to get all screeners
 */
export function useScreeners() {
  return useQuery({
    queryKey: screenerKeys.lists(),
    queryFn: async () => {
      const response = await screenerService.getAllScreeners();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch screeners");
      }
      return response.data || [];
    },
    ...CACHE_CONFIG.LIST,
    // Keep previous data while fetching new data
    placeholderData: (previousData) => previousData,
    // Structural sharing for better performance
    structuralSharing: true,
  });
}

/**
 * Hook to get screener by ID
 */
export function useScreenerById(id: string, enabled: boolean = true) {
  return useQuery({
    queryKey: screenerKeys.detail(id),
    queryFn: async () => {
      const response = await screenerService.getScreenerById(id);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch screener");
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
 * Hook to get screener by symbol
 */
export function useScreenerBySymbol(symbol: string, enabled: boolean = true) {
  return useQuery({
    queryKey: screenerKeys.bySymbol(symbol),
    queryFn: async () => {
      const response = await screenerService.getScreenerBySymbol(symbol);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch screener");
      }
      return response.data;
    },
    enabled: enabled && !!symbol,
    ...CACHE_CONFIG.DETAIL,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get screeners with filters, sorting, and pagination
 */
export function useScreenersWithFilters(
  filters?: FilterOptions,
  sort?: SortOptions,
  pagination?: PaginationOptions,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: screenerKeys.list(filters, sort, pagination),
    queryFn: async () => {
      const response = await screenerService.getScreenersWithFilters(
        filters,
        sort,
        pagination
      );
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch screeners");
      }
      return response.data;
    },
    enabled,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to search screeners by symbol
 */
export function useSearchScreeners(
  searchTerm: string,
  limit: number = 10,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: screenerKeys.search(searchTerm, limit),
    queryFn: async () => {
      const response = await screenerService.searchScreenersBySymbol(
        searchTerm,
        limit
      );
      if (!response.success) {
        throw new Error(response.message || "Failed to search screeners");
      }
      return response.data || [];
    },
    enabled: enabled && searchTerm.length > 0,
    ...CACHE_CONFIG.SEARCH,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get screeners by price range
 */
export function useScreenersByPriceRange(
  minPrice: number,
  maxPrice: number,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: screenerKeys.priceRange(minPrice, maxPrice),
    queryFn: async () => {
      const response = await screenerService.getScreenersByPriceRange(
        minPrice,
        maxPrice
      );
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch screeners");
      }
      return response.data || [];
    },
    enabled: enabled && minPrice >= 0 && maxPrice > minPrice,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get screeners by volume range
 */
export function useScreenersByVolumeRange(
  minVolume: number,
  maxVolume: number,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: screenerKeys.volumeRange(minVolume, maxVolume),
    queryFn: async () => {
      const response = await screenerService.getScreenersByVolumeRange(
        minVolume,
        maxVolume
      );
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch screeners");
      }
      return response.data || [];
    },
    enabled: enabled && minVolume >= 0 && maxVolume > minVolume,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get top gainers
 * Uses real-time caching to auto-refetch every 30 seconds
 */
export function useTopGainers(
  limit: number = 10,
  enabled: boolean = true,
  options?: {
    refetchInterval?: number | false;
    disableRealtime?: boolean;
  }
) {
  const cacheConfig = options?.disableRealtime
    ? CACHE_CONFIG.LIST
    : {
        ...CACHE_CONFIG.REALTIME,
        refetchInterval:
          options?.refetchInterval !== undefined
            ? options.refetchInterval
            : CACHE_CONFIG.REALTIME.refetchInterval,
      };

  return useQuery({
    queryKey: screenerKeys.topGainers(limit),
    queryFn: async () => {
      const response = await screenerService.getTopGainers(limit);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch top gainers");
      }
      return response.data || [];
    },
    enabled,
    ...cacheConfig,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get most active stocks
 * Uses real-time caching to auto-refetch every 30 seconds
 */
export function useMostActive(
  limit: number = 10,
  enabled: boolean = true,
  options?: {
    refetchInterval?: number | false;
    disableRealtime?: boolean;
  }
) {
  const cacheConfig = options?.disableRealtime
    ? CACHE_CONFIG.LIST
    : {
        ...CACHE_CONFIG.REALTIME,
        refetchInterval:
          options?.refetchInterval !== undefined
            ? options.refetchInterval
            : CACHE_CONFIG.REALTIME.refetchInterval,
      };

  return useQuery({
    queryKey: screenerKeys.mostActive(limit),
    queryFn: async () => {
      const response = await screenerService.getMostActive(limit);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch most active stocks");
      }
      return response.data || [];
    },
    enabled,
    ...cacheConfig,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get screener count
 */
export function useScreenerCount(enabled: boolean = true) {
  return useQuery({
    queryKey: screenerKeys.count(),
    queryFn: async () => {
      const response = await screenerService.getScreenerCount();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch count");
      }
      return response.data?.count || 0;
    },
    enabled,
    ...CACHE_CONFIG.COUNT,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get screeners by multiple symbols (mutation for flexibility)
 */
export function useScreenersBySymbols() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (symbols: string[]) => {
      const response = await screenerService.getScreenersBySymbols(symbols);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch screeners");
      }
      return response.data || [];
    },
    onSuccess: (data, symbols) => {
      // Cache individual screener results
      data.forEach((screener) => {
        queryClient.setQueryData(screenerKeys.detail(screener.id), screener);
        queryClient.setQueryData(screenerKeys.bySymbol(screener.symbol), screener);
      });
      // Cache the symbols query
      queryClient.setQueryData(screenerKeys.bySymbols(symbols), data);
    },
  });
}

/**
 * Hook to invalidate all screener queries
 */
export function useInvalidateScreeners() {
  const queryClient = useQueryClient();

  return () => {
    queryClient.invalidateQueries({ queryKey: screenerKeys.all });
  };
}

/**
 * Prefetch a screener by ID for instant loading
 */
export function usePrefetchScreenerById() {
  const queryClient = useQueryClient();

  return (id: string) => {
    queryClient.prefetchQuery({
      queryKey: screenerKeys.detail(id),
      queryFn: async () => {
        const response = await screenerService.getScreenerById(id);
        if (!response.success) {
          throw new Error(response.message || "Failed to fetch screener");
        }
        return response.data;
      },
      ...CACHE_CONFIG.DETAIL,
    });
  };
}

/**
 * Prefetch a screener by symbol for instant loading
 */
export function usePrefetchScreenerBySymbol() {
  const queryClient = useQueryClient();

  return (symbol: string) => {
    queryClient.prefetchQuery({
      queryKey: screenerKeys.bySymbol(symbol),
      queryFn: async () => {
        const response = await screenerService.getScreenerBySymbol(symbol);
        if (!response.success) {
          throw new Error(response.message || "Failed to fetch screener");
        }
        return response.data;
      },
      ...CACHE_CONFIG.DETAIL,
    });
  };
}

/**
 * Prefetch all screeners list
 */
export function usePrefetchScreeners() {
  const queryClient = useQueryClient();

  return () => {
    queryClient.prefetchQuery({
      queryKey: screenerKeys.lists(),
      queryFn: async () => {
        const response = await screenerService.getAllScreeners();
        if (!response.success) {
          throw new Error(response.message || "Failed to fetch screeners");
        }
        return response.data || [];
      },
      ...CACHE_CONFIG.LIST,
    });
  };
}

/**
 * Get cached screener data without triggering a fetch
 */
export function useGetCachedScreener() {
  const queryClient = useQueryClient();

  return {
    byId: (id: string): Screener | undefined => {
      return queryClient.getQueryData<Screener>(screenerKeys.detail(id));
    },
    bySymbol: (symbol: string): Screener | undefined => {
      return queryClient.getQueryData<Screener>(screenerKeys.bySymbol(symbol));
    },
    list: (): Screener[] | undefined => {
      return queryClient.getQueryData<Screener[]>(screenerKeys.lists());
    },
  };
}

/**
 * Set screener data in cache (useful for optimistic updates)
 */
export function useSetCachedScreener() {
  const queryClient = useQueryClient();

  return {
    byId: (id: string, data: Screener) => {
      queryClient.setQueryData(screenerKeys.detail(id), data);
      // Also update by symbol cache if symbol exists
      if (data.symbol) {
        queryClient.setQueryData(screenerKeys.bySymbol(data.symbol), data);
      }
    },
    bySymbol: (symbol: string, data: Screener) => {
      queryClient.setQueryData(screenerKeys.bySymbol(symbol), data);
      // Also update by id cache
      queryClient.setQueryData(screenerKeys.detail(data.id), data);
    },
  };
}


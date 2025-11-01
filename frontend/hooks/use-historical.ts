/**
 * TanStack Query hooks for Historical API
 * Provides reactive data fetching with caching and error handling
 */

import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import type {
  Historical,
  HistoricalFilterOptions,
  HistoricalPaginationOptions,
  HistoricalSortOptions,
  VolumeMetricsResult,
} from "../lib/types/historical";
import * as historicalService from "../lib/service/historical.service";

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
  // Volume metrics - cache for 10 minutes, stale after 5 minutes
  METRICS: {
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
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
export const historicalKeys = {
  all: ["historical"] as const,
  lists: () => [...historicalKeys.all, "list"] as const,
  list: (
    filters?: HistoricalFilterOptions,
    sort?: HistoricalSortOptions,
    pagination?: HistoricalPaginationOptions
  ) => [...historicalKeys.lists(), filters, sort, pagination] as const,
  details: () => [...historicalKeys.all, "detail"] as const,
  detail: (id: string) => [...historicalKeys.details(), id] as const,
  bySymbol: (symbol: string) =>
    [...historicalKeys.all, "symbol", symbol] as const,
  bySymbolAndParams: (symbol: string, range: string, interval: string) =>
    [...historicalKeys.all, "symbol", symbol, "params", range, interval] as const,
  count: () => [...historicalKeys.all, "count"] as const,
  countBySymbol: (symbol: string) =>
    [...historicalKeys.all, "count", symbol] as const,
  volumeMetrics: () => [...historicalKeys.all, "volume-metrics"] as const,
};

/**
 * Hook to get all historical records
 */
export function useHistorical() {
  return useQuery({
    queryKey: historicalKeys.lists(),
    queryFn: async () => {
      const response = await historicalService.getAllHistorical();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch historical records");
      }
      return response.data || [];
    },
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get historical record by ID
 */
export function useHistoricalById(id: string, enabled: boolean = true) {
  return useQuery({
    queryKey: historicalKeys.detail(id),
    queryFn: async () => {
      const response = await historicalService.getHistoricalById(id);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch historical record");
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
 * Hook to get historical records by symbol
 */
export function useHistoricalBySymbol(symbol: string, enabled: boolean = true) {
  return useQuery({
    queryKey: historicalKeys.bySymbol(symbol),
    queryFn: async () => {
      const response = await historicalService.getHistoricalBySymbol(symbol);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch historical records");
      }
      return response.data || [];
    },
    enabled: enabled && !!symbol,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get historical records by symbol, range, and interval
 */
export function useHistoricalBySymbolAndParams(
  symbol: string,
  range: string,
  interval: string,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.bySymbolAndParams(symbol, range, interval),
    queryFn: async () => {
      const response = await historicalService.getHistoricalBySymbolAndParams(
        symbol,
        range,
        interval
      );
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch historical records");
      }
      return response.data || [];
    },
    enabled: enabled && !!symbol && !!range && !!interval,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get historical records with filters, sorting, and pagination
 */
export function useHistoricalWithFilters(
  filters?: HistoricalFilterOptions,
  sort?: HistoricalSortOptions,
  pagination?: HistoricalPaginationOptions,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.list(filters, sort, pagination),
    queryFn: async () => {
      const response = await historicalService.getHistoricalWithFilters(
        filters,
        sort,
        pagination
      );
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch historical records");
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
 * Hook to get historical count
 */
export function useHistoricalCount(enabled: boolean = true) {
  return useQuery({
    queryKey: historicalKeys.count(),
    queryFn: async () => {
      const response = await historicalService.getHistoricalCount();
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
 * Hook to get historical count by symbol
 */
export function useHistoricalCountBySymbol(
  symbol: string,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.countBySymbol(symbol),
    queryFn: async () => {
      const response = await historicalService.getHistoricalCountBySymbol(symbol);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch count");
      }
      return response.data?.count || 0;
    },
    enabled: enabled && !!symbol,
    ...CACHE_CONFIG.COUNT,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get stocks volume metrics
 */
export function useStocksVolumeMetrics(enabled: boolean = true) {
  return useQuery({
    queryKey: historicalKeys.volumeMetrics(),
    queryFn: async () => {
      const response = await historicalService.getStocksVolumeMetrics();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch volume metrics");
      }
      return response.data || [];
    },
    enabled,
    ...CACHE_CONFIG.METRICS,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to create a historical record
 */
export function useCreateHistorical() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (historical: Historical) => {
      const response = await historicalService.createHistorical(historical);
      if (!response.success) {
        throw new Error(response.message || "Failed to create historical record");
      }
      return response.data;
    },
    onSuccess: (data) => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: historicalKeys.all });
      // Cache the new record
      queryClient.setQueryData(historicalKeys.detail(data.id), data);
    },
  });
}

/**
 * Hook to create historical records in batch
 */
export function useCreateHistoricalBatch() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (historical: Historical[]) => {
      const response = await historicalService.createHistoricalBatch(historical);
      if (!response.success) {
        throw new Error(response.message || "Failed to create historical records");
      }
      return response.data;
    },
    onSuccess: () => {
      // Invalidate all historical queries
      queryClient.invalidateQueries({ queryKey: historicalKeys.all });
    },
  });
}

/**
 * Hook to upsert a historical record
 */
export function useUpsertHistorical() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (historical: Historical) => {
      const response = await historicalService.upsertHistorical(historical);
      if (!response.success) {
        throw new Error(response.message || "Failed to upsert historical record");
      }
      return response.data;
    },
    onSuccess: (data) => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: historicalKeys.all });
      // Cache the upserted record
      queryClient.setQueryData(historicalKeys.detail(data.id), data);
    },
  });
}

/**
 * Hook to upsert historical records in batch
 */
export function useUpsertHistoricalBatch() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (historical: Historical[]) => {
      const response = await historicalService.upsertHistoricalBatch(historical);
      if (!response.success) {
        throw new Error(response.message || "Failed to upsert historical records");
      }
      return response.data;
    },
    onSuccess: () => {
      // Invalidate all historical queries
      queryClient.invalidateQueries({ queryKey: historicalKeys.all });
    },
  });
}

/**
 * Hook to update a historical record
 */
export function useUpdateHistorical() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      id,
      historical,
    }: {
      id: string;
      historical: Partial<Historical>;
    }) => {
      const response = await historicalService.updateHistorical(id, historical);
      if (!response.success) {
        throw new Error(response.message || "Failed to update historical record");
      }
      return response.data;
    },
    onSuccess: (_, variables) => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: historicalKeys.all });
      queryClient.invalidateQueries({
        queryKey: historicalKeys.detail(variables.id),
      });
    },
  });
}

/**
 * Hook to delete a historical record
 */
export function useDeleteHistorical() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await historicalService.deleteHistorical(id);
      if (!response.success) {
        throw new Error(response.message || "Failed to delete historical record");
      }
      return response.data;
    },
    onSuccess: (_, id) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: historicalKeys.detail(id) });
      // Invalidate list queries
      queryClient.invalidateQueries({ queryKey: historicalKeys.lists() });
    },
  });
}

/**
 * Hook to delete historical records by symbol
 */
export function useDeleteHistoricalBySymbol() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (symbol: string) => {
      const response = await historicalService.deleteHistoricalBySymbol(symbol);
      if (!response.success) {
        throw new Error(response.message || "Failed to delete historical records");
      }
      return response.data;
    },
    onSuccess: (_, symbol) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: historicalKeys.bySymbol(symbol) });
      // Invalidate list queries
      queryClient.invalidateQueries({ queryKey: historicalKeys.lists() });
    },
  });
}

/**
 * Hook to delete historical records by symbol, range, and interval
 */
export function useDeleteHistoricalBySymbolAndParams() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      symbol,
      range,
      interval,
    }: {
      symbol: string;
      range: string;
      interval: string;
    }) => {
      const response = await historicalService.deleteHistoricalBySymbolAndParams(
        symbol,
        range,
        interval
      );
      if (!response.success) {
        throw new Error(response.message || "Failed to delete historical records");
      }
      return response.data;
    },
    onSuccess: (_, variables) => {
      // Remove from cache
      queryClient.removeQueries({
        queryKey: historicalKeys.bySymbolAndParams(
          variables.symbol,
          variables.range,
          variables.interval
        ),
      });
      // Invalidate list queries
      queryClient.invalidateQueries({ queryKey: historicalKeys.lists() });
    },
  });
}

/**
 * Hook to invalidate all historical queries
 */
export function useInvalidateHistorical() {
  const queryClient = useQueryClient();

  return () => {
    queryClient.invalidateQueries({ queryKey: historicalKeys.all });
  };
}

/**
 * Prefetch a historical record by ID for instant loading
 */
export function usePrefetchHistoricalById() {
  const queryClient = useQueryClient();

  return (id: string) => {
    queryClient.prefetchQuery({
      queryKey: historicalKeys.detail(id),
      queryFn: async () => {
        const response = await historicalService.getHistoricalById(id);
        if (!response.success) {
          throw new Error(response.message || "Failed to fetch historical record");
        }
        return response.data;
      },
      ...CACHE_CONFIG.DETAIL,
    });
  };
}

/**
 * Prefetch historical records by symbol for instant loading
 */
export function usePrefetchHistoricalBySymbol() {
  const queryClient = useQueryClient();

  return (symbol: string) => {
    queryClient.prefetchQuery({
      queryKey: historicalKeys.bySymbol(symbol),
      queryFn: async () => {
        const response = await historicalService.getHistoricalBySymbol(symbol);
        if (!response.success) {
          throw new Error(response.message || "Failed to fetch historical records");
        }
        return response.data || [];
      },
      ...CACHE_CONFIG.LIST,
    });
  };
}

/**
 * Get cached historical data without triggering a fetch
 */
export function useGetCachedHistorical() {
  const queryClient = useQueryClient();

  return {
    byId: (id: string): Historical | undefined => {
      return queryClient.getQueryData<Historical>(historicalKeys.detail(id));
    },
    bySymbol: (symbol: string): Historical[] | undefined => {
      return queryClient.getQueryData<Historical[]>(historicalKeys.bySymbol(symbol));
    },
    list: (): Historical[] | undefined => {
      return queryClient.getQueryData<Historical[]>(historicalKeys.lists());
    },
    volumeMetrics: (): VolumeMetricsResult[] | undefined => {
      return queryClient.getQueryData<VolumeMetricsResult[]>(
        historicalKeys.volumeMetrics()
      );
    },
  };
}


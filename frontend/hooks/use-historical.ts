/**
 * TanStack Query hooks for Historical API
 * Provides reactive data fetching with caching and error handling
 */

import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import type { Historical } from "../lib/types/historical";
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
  // Screening queries - cache for 2 minutes, stale after 1 minute
  SCREENING: {
    staleTime: 1 * 60 * 1000, // 1 minute
    gcTime: 2 * 60 * 1000, // 2 minutes
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
  details: () => [...historicalKeys.all, "detail"] as const,
  detail: (id: string) => [...historicalKeys.details(), id] as const,
  bySymbol: (symbol: string, range: string, interval: string) =>
    [...historicalKeys.all, "by-symbol", symbol, range, interval] as const,
  // Screening query keys
  screening: () => [...historicalKeys.all, "screening"] as const,
  insideDay: () => [...historicalKeys.screening(), "inside-day"] as const,
  highVolumeQuarter: () => [...historicalKeys.screening(), "high-volume-quarter"] as const,
  highVolumeYear: () => [...historicalKeys.screening(), "high-volume-year"] as const,
  highVolumeEver: () => [...historicalKeys.screening(), "high-volume-ever"] as const,
  adrScreen: (params?: { range?: string; interval?: string; lookback?: number; minAdr?: number; maxAdr?: number }) =>
    [...historicalKeys.screening(), "adr-screen", params] as const,
  atrScreen: (params?: { range?: string; interval?: string; lookback?: number; minAtr?: number; maxAtr?: number }) =>
    [...historicalKeys.screening(), "atr-screen", params] as const,
  adrForSymbol: (symbol: string, params?: { range?: string; interval?: string; lookback?: number }) =>
    [...historicalKeys.screening(), "adr", symbol, params] as const,
  atrForSymbol: (symbol: string, params?: { range?: string; interval?: string; lookback?: number }) =>
    [...historicalKeys.screening(), "atr", symbol, params] as const,
  avgVolumeDollarsScreen: (params?: { range?: string; interval?: string; lookback?: number; minVolDollarsM?: number; maxVolDollarsM?: number }) =>
    [...historicalKeys.screening(), "avg-volume-dollars-screen", params] as const,
  avgVolumePercentScreen: (params?: { range?: string; interval?: string; lookback?: number; minVolPercent?: number; maxVolPercent?: number }) =>
    [...historicalKeys.screening(), "avg-volume-percent-screen", params] as const,
  avgVolumeDollarsForSymbol: (symbol: string, params?: { range?: string; interval?: string; lookback?: number }) =>
    [...historicalKeys.screening(), "avg-volume-dollars", symbol, params] as const,
  avgVolumePercentForSymbol: (symbol: string, params?: { range?: string; interval?: string; lookback?: number }) =>
    [...historicalKeys.screening(), "avg-volume-percent", symbol, params] as const,
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
 * Hook to get historical records by symbol, range, and interval
 */
export function useHistoricalBySymbol(
  symbol: string,
  range: string,
  interval: string,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.bySymbol(symbol, range, interval),
    queryFn: async () => {
      const response = await historicalService.getHistoricalBySymbol(
        symbol,
        range,
        interval
      );
      if (!response.success) {
        throw new Error(
          response.message || "Failed to fetch historical records"
        );
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
 * Public Screening Hooks
 */

export function useInsideDaySymbols(enabled: boolean = true) {
  return useQuery({
    queryKey: historicalKeys.insideDay(),
    queryFn: async () => {
      const response = await historicalService.getInsideDaySymbols();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch inside day symbols");
      }
      return response.data;
    },
    enabled,
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useHighVolumeQuarterSymbols(enabled: boolean = true) {
  return useQuery({
    queryKey: historicalKeys.highVolumeQuarter(),
    queryFn: async () => {
      const response = await historicalService.getHighVolumeQuarterSymbols();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch high volume quarter symbols");
      }
      return response.data;
    },
    enabled,
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useHighVolumeYearSymbols(enabled: boolean = true) {
  return useQuery({
    queryKey: historicalKeys.highVolumeYear(),
    queryFn: async () => {
      const response = await historicalService.getHighVolumeYearSymbols();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch high volume year symbols");
      }
      return response.data;
    },
    enabled,
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useHighVolumeEverSymbols(enabled: boolean = true) {
  return useQuery({
    queryKey: historicalKeys.highVolumeEver(),
    queryFn: async () => {
      const response = await historicalService.getHighVolumeEverSymbols();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch high volume ever symbols");
      }
      return response.data;
    },
    enabled,
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useAdrScreen(
  params: {
    range?: string;
    interval?: string;
    lookback?: number;
    minAdr?: number;
    maxAdr?: number;
  },
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.adrScreen(params),
    queryFn: async () => {
      const response = await historicalService.getAdrScreen(params);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch ADR screen");
      }
      return response.data;
    },
    enabled: enabled && !!(params.range && params.interval),
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useAtrScreen(
  params: {
    range?: string;
    interval?: string;
    lookback?: number;
    minAtr?: number;
    maxAtr?: number;
  },
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.atrScreen(params),
    queryFn: async () => {
      const response = await historicalService.getAtrScreen(params);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch ATR screen");
      }
      return response.data;
    },
    enabled: enabled && !!(params.range && params.interval),
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useAdrForSymbol(
  params: {
    symbol: string;
    range?: string;
    interval?: string;
    lookback?: number;
  },
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.adrForSymbol(params.symbol, params),
    queryFn: async () => {
      const response = await historicalService.getAdrForSymbol(params);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch ADR for symbol");
      }
      return response.data;
    },
    enabled: enabled && !!params.symbol && !!params.range && !!params.interval,
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useAtrForSymbol(
  params: {
    symbol: string;
    range?: string;
    interval?: string;
    lookback?: number;
  },
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.atrForSymbol(params.symbol, params),
    queryFn: async () => {
      const response = await historicalService.getAtrForSymbol(params);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch ATR for symbol");
      }
      return response.data;
    },
    enabled: enabled && !!params.symbol && !!params.range && !!params.interval,
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useAvgVolumeDollarsScreen(
  params: {
    range?: string;
    interval?: string;
    lookback?: number;
    minVolDollarsM?: number;
    maxVolDollarsM?: number;
  },
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.avgVolumeDollarsScreen(params),
    queryFn: async () => {
      const response = await historicalService.getAvgVolumeDollarsScreen(params);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch avg volume dollars screen");
      }
      return response.data;
    },
    enabled: enabled && !!(params.range && params.interval),
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useAvgVolumePercentScreen(
  params: {
    range?: string;
    interval?: string;
    lookback?: number;
    minVolPercent?: number;
    maxVolPercent?: number;
  },
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.avgVolumePercentScreen(params),
    queryFn: async () => {
      const response = await historicalService.getAvgVolumePercentScreen(params);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch avg volume percent screen");
      }
      return response.data;
    },
    enabled: enabled && !!(params.range && params.interval),
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useAvgVolumeDollarsForSymbol(
  params: {
    symbol: string;
    range?: string;
    interval?: string;
    lookback?: number;
  },
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.avgVolumeDollarsForSymbol(params.symbol, params),
    queryFn: async () => {
      const response = await historicalService.getAvgVolumeDollarsForSymbol(params);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch avg volume dollars for symbol");
      }
      return response.data;
    },
    enabled: enabled && !!params.symbol && !!params.range && !!params.interval,
    ...CACHE_CONFIG.SCREENING,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

export function useAvgVolumePercentForSymbol(
  params: {
    symbol: string;
    range?: string;
    interval?: string;
    lookback?: number;
  },
  enabled: boolean = true
) {
  return useQuery({
    queryKey: historicalKeys.avgVolumePercentForSymbol(params.symbol, params),
    queryFn: async () => {
      const response = await historicalService.getAvgVolumePercentForSymbol(params);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch avg volume percent for symbol");
      }
      return response.data;
    },
    enabled: enabled && !!params.symbol && !!params.range && !!params.interval,
    ...CACHE_CONFIG.SCREENING,
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
 * Get cached historical data without triggering a fetch
 */
export function useGetCachedHistorical() {
  const queryClient = useQueryClient();

  return {
    byId: (id: string): Historical | undefined => {
      return queryClient.getQueryData<Historical>(historicalKeys.detail(id));
    },
    list: (): Historical[] | undefined => {
      return queryClient.getQueryData<Historical[]>(historicalKeys.lists());
    },
  };
}


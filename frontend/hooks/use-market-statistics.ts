/**
 * TanStack Query hooks for Market Statistics API
 * Provides reactive data fetching with caching and error handling
 */

import { useQuery } from "@tanstack/react-query";
import * as marketStatisticsService from "../lib/service/market-statistics.service";

/**
 * Cache configuration constants
 */
const CACHE_CONFIG = {
  // Historical data - cache for 10 minutes, stale after 5 minutes
  HISTORICAL: {
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
  },
  // Current stats - cache for 1 minute, stale after 30 seconds, refetch every 5 minutes
  CURRENT: {
    staleTime: 30 * 1000, // 30 seconds
    gcTime: 1 * 60 * 1000, // 1 minute
    refetchInterval: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
  },
} as const;

/**
 * Query keys for TanStack Query
 */
export const marketStatisticsKeys = {
  all: ["market-statistics"] as const,
  historical: (startDate?: string, endDate?: string) =>
    [...marketStatisticsKeys.all, "historical", startDate, endDate] as const,
  current: () => [...marketStatisticsKeys.all, "current"] as const,
};

/**
 * Hook to get historical market statistics for charting
 */
export function useMarketStatistics(
  startDate?: string,
  endDate?: string,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: marketStatisticsKeys.historical(startDate, endDate),
    queryFn: async () => {
      const response = await marketStatisticsService.getMarketStatistics(
        startDate,
        endDate
      );
      if (!response.success) {
        throw new Error("Failed to fetch market statistics");
      }
      return response.data || [];
    },
    enabled,
    ...CACHE_CONFIG.HISTORICAL,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get today's real-time market statistics
 */
export function useCurrentMarketStatistics(enabled: boolean = true) {
  return useQuery({
    queryKey: marketStatisticsKeys.current(),
    queryFn: async () => {
      const response =
        await marketStatisticsService.getCurrentMarketStatistics();
      if (!response.success) {
        throw new Error("Failed to fetch current market statistics");
      }
      return response.data;
    },
    enabled,
    ...CACHE_CONFIG.CURRENT,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}


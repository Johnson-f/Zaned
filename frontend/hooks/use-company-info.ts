/**
 * TanStack Query hooks for Company Info API
 * Provides reactive data fetching with caching and error handling
 */

import {
  useQuery,
} from "@tanstack/react-query";
import * as companyInfoService from "../lib/service/company-info.service";

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
  // Search queries - cache for 1 minute, stale immediately
  SEARCH: {
    staleTime: 0, // Stale immediately for search results
    gcTime: 1 * 60 * 1000, // 1 minute
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  },
} as const;

/**
 * Query keys for TanStack Query
 */
export const companyInfoKeys = {
  all: ["company-info"] as const,
  lists: () => [...companyInfoKeys.all, "list"] as const,
  details: () => [...companyInfoKeys.all, "detail"] as const,
  detail: (symbol: string) => [...companyInfoKeys.details(), symbol] as const,
  search: (query: string) => [...companyInfoKeys.all, "search", query] as const,
  bySector: (sector: string) => [...companyInfoKeys.all, "sector", sector] as const,
  byIndustry: (industry: string) => [...companyInfoKeys.all, "industry", industry] as const,
};

/**
 * Hook to get all company info
 */
export function useCompanyInfo(enabled: boolean = true) {
  return useQuery({
    queryKey: companyInfoKeys.lists(),
    queryFn: async () => {
      const response = await companyInfoService.getAllCompanyInfo();
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch company info");
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
 * Hook to get company info by symbol
 */
export function useCompanyInfoBySymbol(symbol: string, enabled: boolean = true) {
  return useQuery({
    queryKey: companyInfoKeys.detail(symbol),
    queryFn: async () => {
      const response = await companyInfoService.getCompanyInfoBySymbol(symbol);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch company info");
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
 * Hook to search company info
 */
export function useSearchCompanyInfo(
  searchTerm: string,
  enabled: boolean = true
) {
  return useQuery({
    queryKey: companyInfoKeys.search(searchTerm),
    queryFn: async () => {
      const response = await companyInfoService.searchCompanyInfo(searchTerm);
      if (!response.success) {
        throw new Error(response.message || "Failed to search company info");
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
 * Hook to get company info by sector
 */
export function useCompanyInfoBySector(sector: string, enabled: boolean = true) {
  return useQuery({
    queryKey: companyInfoKeys.bySector(sector),
    queryFn: async () => {
      const response = await companyInfoService.getCompanyInfoBySector(sector);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch company info");
      }
      return response.data || [];
    },
    enabled: enabled && !!sector,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}

/**
 * Hook to get company info by industry
 */
export function useCompanyInfoByIndustry(industry: string, enabled: boolean = true) {
  return useQuery({
    queryKey: companyInfoKeys.byIndustry(industry),
    queryFn: async () => {
      const response = await companyInfoService.getCompanyInfoByIndustry(industry);
      if (!response.success) {
        throw new Error(response.message || "Failed to fetch company info");
      }
      return response.data || [];
    },
    enabled: enabled && !!industry,
    ...CACHE_CONFIG.LIST,
    placeholderData: (previousData) => previousData,
    structuralSharing: true,
  });
}


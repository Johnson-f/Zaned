/**
 * Company Info Service
 * Handles all API calls to the backend company info endpoints
 */

import { API_ENDPOINTS } from "../config/api";
import type {
  ApiResponse,
  CompanyInfo,
} from "../types/company-info";

/**
 * Base fetch function with error handling (public endpoint, no auth needed)
 */
async function fetchApi<T>(
  url: string,
  options?: RequestInit
): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
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
 * Get all company info
 */
export async function getAllCompanyInfo(): Promise<ApiResponse<CompanyInfo[]>> {
  return fetchApi<CompanyInfo[]>(API_ENDPOINTS.COMPANY_INFO.BASE);
}

/**
 * Get company info by symbol
 */
export async function getCompanyInfoBySymbol(
  symbol: string
): Promise<ApiResponse<CompanyInfo>> {
  return fetchApi<CompanyInfo>(API_ENDPOINTS.COMPANY_INFO.BY_SYMBOL(symbol));
}

/**
 * Get company info by multiple symbols
 */
export async function getCompanyInfoBySymbols(
  symbols: string[]
): Promise<ApiResponse<CompanyInfo[]>> {
  return fetchApi<CompanyInfo[]>(API_ENDPOINTS.COMPANY_INFO.BY_SYMBOLS, {
    method: "POST",
    body: JSON.stringify({ symbols }),
  });
}

/**
 * Search company info by name, sector, industry, or symbol
 */
export async function searchCompanyInfo(
  query: string
): Promise<ApiResponse<CompanyInfo[]>> {
  return fetchApi<CompanyInfo[]>(API_ENDPOINTS.COMPANY_INFO.SEARCH(query));
}

/**
 * Get company info by sector
 */
export async function getCompanyInfoBySector(
  sector: string
): Promise<ApiResponse<CompanyInfo[]>> {
  return fetchApi<CompanyInfo[]>(API_ENDPOINTS.COMPANY_INFO.BY_SECTOR(sector));
}

/**
 * Get company info by industry
 */
export async function getCompanyInfoByIndustry(
  industry: string
): Promise<ApiResponse<CompanyInfo[]>> {
  return fetchApi<CompanyInfo[]>(API_ENDPOINTS.COMPANY_INFO.BY_INDUSTRY(industry));
}


package caching

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const cachePrefix = "cache"

// GenerateKey generates a cache key from endpoint path and query parameters
func GenerateKey(endpoint string, params map[string]string) string {
	// Normalize endpoint (remove leading/trailing slashes)
	endpoint = strings.Trim(endpoint, "/")
	
	// If no params, return simple key
	if len(params) == 0 {
		return fmt.Sprintf("%s:%s", cachePrefix, endpoint)
	}
	
	// Sort params for consistent key generation
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Build query string from sorted params
	var queryParts []string
	for _, k := range keys {
		v := params[k]
		if v != "" {
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
		}
	}
	
	queryString := strings.Join(queryParts, "&")
	
	// Hash long query strings to keep keys manageable
	if len(queryString) > 100 {
		hash := sha256.Sum256([]byte(queryString))
		queryString = hex.EncodeToString(hash[:])[:16] // Use first 16 chars of hash
	}
	
	return fmt.Sprintf("%s:%s:%s", cachePrefix, endpoint, queryString)
}

// GenerateKeyFromPath generates a cache key from a full path (e.g., "/api/company-info/AAPL")
func GenerateKeyFromPath(path string) string {
	path = strings.Trim(path, "/")
	// Remove /api prefix if present
	path = strings.TrimPrefix(path, "api/")
	return fmt.Sprintf("%s:%s", cachePrefix, path)
}

// GenerateKeyFromQuery generates a cache key from endpoint and query string
func GenerateKeyFromQuery(endpoint string, queryString string) string {
	endpoint = strings.Trim(endpoint, "/")
	
	if queryString == "" {
		return fmt.Sprintf("%s:%s", cachePrefix, endpoint)
	}
	
	// Parse and normalize query string
	values, err := url.ParseQuery(queryString)
	if err != nil {
		// If parsing fails, hash the raw query string
		hash := sha256.Sum256([]byte(queryString))
		return fmt.Sprintf("%s:%s:%s", cachePrefix, endpoint, hex.EncodeToString(hash[:])[:16])
	}
	
	// Convert to map for consistent key generation
	params := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			params[k] = v[0] // Take first value
		}
	}
	
	return GenerateKey(endpoint, params)
}

// GeneratePattern generates a pattern for matching multiple cache keys
// Useful for invalidation (e.g., "cache:company-info:*")
func GeneratePattern(endpoint string) string {
	endpoint = strings.Trim(endpoint, "/")
	return fmt.Sprintf("%s:%s:*", cachePrefix, endpoint)
}
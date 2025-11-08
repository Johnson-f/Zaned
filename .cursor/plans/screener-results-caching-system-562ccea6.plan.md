<!-- 562ccea6-711c-4b7a-8008-777ac336de85 6329138f-78a2-452a-9d47-5574198ea563 -->
# Screener Results Caching System

## Overview

Replace real-time computation endpoints with a database-backed caching system that stores screener results daily and allows time period filtering (7d, 30d, 90d, ytd, all).

## Backend Changes

### 1. Create Database Model

**File:** `backend/model/screener-result.go`

- Create `ScreenerResult` model with fields: `id`, `type` (inside_day, high_volume_quarter, high_volume_year, high_volume_ever), `symbol`, `date`, timestamps
- Add unique constraint on (type, symbol, date)
- Add indexes on type, symbol, and date for query performance

### 2. Add Service Methods

**File:** `backend/service/historical.go`

- Add `SaveInsideDayResults()` - calls `GetSymbolsWithDailyInsideDay()` and saves to DB
- Add `SaveHighVolumeQuarterResults()` - calls `GetSymbolsWithHighestVolumeInQuarter()` and saves to DB
- Add `SaveHighVolumeYearResults()` - calls `GetSymbolsWithHighestVolumeInYear()` and saves to DB
- Add `SaveHighVolumeEverResults()` - calls `GetSymbolsWithHighestVolumeEver()` and saves to DB
- Add `GetScreenerResults(resultType, period)` - fetches from DB with period filtering (7d, 30d, 90d, ytd, all)
- All save methods use upsert with ON CONFLICT to prevent duplicates

### 3. Add Admin Endpoints (for cron jobs)

**File:** `backend/routes/routes.go`

- `POST /api/admin/screener/save-inside-day` - triggers `SaveInsideDayResults()`
- `POST /api/admin/screener/save-high-volume-quarter` - triggers `SaveHighVolumeQuarterResults()`
- `POST /api/admin/screener/save-high-volume-year` - triggers `SaveHighVolumeYearResults()`
- `POST /api/admin/screener/save-high-volume-ever` - triggers `SaveHighVolumeEverResults()`

### 4. Add Public Endpoint (for frontend)

**File:** `backend/routes/routes.go`

- `GET /api/screener-results?type={type}&period={period}` - fetches cached results
- `type`: inside_day, high_volume_quarter, high_volume_year, high_volume_ever
- `period`: 7d, 30d, 90d, ytd, all (default: all)
- Returns: `{ success: true, data: { symbols: string[], count: number, type: string, period: string } }`

### 5. Remove Old Endpoints

**File:** `backend/routes/routes.go`

- Remove `GET /api/inside-day` (lines 249-267)
- Remove `GET /api/high-volume-quarter` (lines 269-287)
- Remove `GET /api/high-volume-year` (lines 289-307)
- Remove `GET /api/high-volume-ever` (lines 309-327)

### 6. Update Database Migration

**File:** `backend/main.go` or migration file

- Add `&model.ScreenerResult{}` to the migration list

## Frontend Changes

### 7. Update API Configuration

**File:** `frontend/lib/config/api.ts`

- Remove from `SCREENING` object:
- `INSIDE_DAY`
- `HIGH_VOLUME_QUARTER`
- `HIGH_VOLUME_YEAR`
- `HIGH_VOLUME_EVER`
- Add to `ADMIN` object:
- `SAVE_INSIDE_DAY: ${API_BASE_URL}/api/admin/screener/save-inside-day`
- `SAVE_HIGH_VOLUME_QUARTER: ${API_BASE_URL}/api/admin/screener/save-high-volume-quarter`
- `SAVE_HIGH_VOLUME_YEAR: ${API_BASE_URL}/api/admin/screener/save-high-volume-year`
- `SAVE_HIGH_VOLUME_EVER: ${API_BASE_URL}/api/admin/screener/save-high-volume-ever`
- Add new `SCREENER_RESULTS` object:
- `BASE: (type: string, period?: string) => ${API_BASE_URL}/api/screener-results?type=${type}&period=${period || 'all'}`

### 8. Update Service

**File:** `frontend/lib/service/historical.service.ts`

- Remove functions:
- `getInsideDaySymbols()`
- `getHighVolumeQuarterSymbols()`
- `getHighVolumeYearSymbols()`
- `getHighVolumeEverSymbols()`
- Add new function:
- `getScreenerResults(type: string, period?: string)` - calls new endpoint

### 9. Update Types

**File:** `frontend/lib/types/historical.ts` (if exists) or create new

- Add `ScreenerResultType` type: `"inside_day" | "high_volume_quarter" | "high_volume_year" | "high_volume_ever"`
- Add `ScreenerResultPeriod` type: `"7d" | "30d" | "90d" | "ytd" | "all"`
- Add `ScreenerResultsResponse` interface with `symbols`, `count`, `type`, `period`

### 10. Update Hooks

**File:** `frontend/hooks/use-historical.ts`

- Remove hooks:
- `useInsideDaySymbols()`
- `useHighVolumeQuarterSymbols()`
- `useHighVolumeYearSymbols()`
- `useHighVolumeEverSymbols()`
- Remove query keys:
- `insideDay()`
- `highVolumeQuarter()`
- `highVolumeYear()`
- `highVolumeEver()`
- Add new hook:
- `useScreenerResults(type: ScreenerResultType, period?: ScreenerResultPeriod, enabled?: boolean)` - uses new endpoint with caching

### 11. Update Components

**Files:**

- `frontend/components/watchlist/inside-day-watchlist.tsx`
- `frontend/components/watchlist/highest-volume-quarter.tsx`
- `frontend/components/watchlist/highest-volume-year.tsx`
- `frontend/components/watchlist/highest-volume-ever.tsx`
- `frontend/components/watchlist/index.tsx`

- Replace `useInsideDaySymbols()` with `useScreenerResults("inside_day", "all")`
- Replace `useHighVolumeQuarterSymbols()` with `useScreenerResults("high_volume_quarter", "all")`
- Replace `useHighVolumeYearSymbols()` with `useScreenerResults("high_volume_year", "all")`
- Replace `useHighVolumeEverSymbols()` with `useScreenerResults("high_volume_ever", "all")`
- Update data access from `data.symbols` to match new response structure
- Add period selector UI (optional enhancement) to allow users to filter by time period

## Testing & Validation

### 12. Verify Database Migration

- Ensure `screener_results` table is created with proper indexes and constraints

### 13. Test Admin Endpoints

- Verify all 4 save endpoints work correctly
- Verify upsert prevents duplicates
- Verify data is saved with correct date

### 14. Test Public Endpoint

- Test all 4 result types
- Test all 5 period options (7d, 30d, 90d, ytd, all)
- Verify response structure matches frontend expectations

### 15. Test Frontend Integration

- Verify all watchlist components load correctly
- Verify symbol counts display correctly
- Verify no broken imports or missing hooks

## Deployment Notes

- Run database migration before deploying
- Set up cron jobs to call admin endpoints daily (recommended: after market close)
- Old endpoints can be removed after confirming new system works - remove this after the implementation
- Consider keeping old endpoints for 1-2 weeks as fallback during transition

### To-dos

- [ ] Create ScreenerResult database model with type, symbol, date fields and unique constraint
- [ ] Add SaveInsideDayResults, SaveHighVolumeQuarterResults, SaveHighVolumeYearResults, SaveHighVolumeEverResults methods to historical service
- [ ] Add GetScreenerResults method with period filtering (7d, 30d, 90d, ytd, all) to historical service
- [ ] Add 4 admin POST endpoints for saving screener results (save-inside-day, save-high-volume-quarter, save-high-volume-year, save-high-volume-ever)
- [ ] Add public GET /api/screener-results endpoint with type and period query parameters
- [ ] Remove old GET endpoints (/inside-day, /high-volume-quarter, /high-volume-year, /high-volume-ever) from routes
- [ ] Update database migration to include ScreenerResult model
- [ ] Update frontend API config: remove old SCREENING endpoints, add new SCREENER_RESULTS endpoint and admin save endpoints
- [ ] Update frontend service: remove old get functions, add getScreenerResults function
- [ ] Update frontend types: add ScreenerResultType, ScreenerResultPeriod, and ScreenerResultsResponse
- [ ] Update frontend hooks: remove old hooks and query keys, add useScreenerResults hook
- [ ] Update all watchlist components to use new useScreenerResults hook instead of old hooks
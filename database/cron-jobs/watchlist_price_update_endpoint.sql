-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/watchlist/update-prices endpoint
--
-- Purpose: Updates price data for all unique stocks in user watchlists. Fetches current prices,
--          after-hours prices, change amounts, and percent changes for all symbols in the
--          watchlist_items table and updates the corresponding records. This ensures users
--          see up-to-date pricing information for stocks they're tracking.
--
-- Schedule: Runs every 15 minutes, Monday to Friday, between 3PM-10PM WAT (15-22 UTC)
--           (*/15 15-22 * * 1-5)
--
-- Actual job
SELECT cron.schedule(
  'watchlist-price-update-15m',
  '*/15 15-22 * * 1-5',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/watchlist/update-prices',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb
    );
  $$
);


--- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/watchlist/update-prices',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 60000
);


-- Endpoints i'm doing
-- https://zaned-backennd.onrender.com/api/admin/ingest/historicals?concurrency=8 -> Update every 4 hours

-- https://zaned-backennd.onrender.com/api/admin/market-statistics/aggregate -> update every 5 minutes monday to friday between 3PMto 10PM


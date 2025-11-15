-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/market-statistics/store-eod endpoint
--
-- Purpose: Stores end-of-day market statistics to the database. This aggregates the daily
--          market statistics (stocks up, down, unchanged counts) that were collected throughout
--          the trading day and saves them as a permanent record for historical analysis.
--          The data is saved to Redis cache first, then persisted to the database.
--
-- Schedule: Runs daily at 8:30 PM UTC (20:30 UTC) - after market close
--           (30 20 * * *)

-- Actual job
SELECT cron.schedule(
  'market-statistics-eod-daily',
  '30 20 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/market-statistics/store-eod',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb,
      timeout_milliseconds := 60000
    );
  $$
);

-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/market-statistics/store-eod',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 60000
);


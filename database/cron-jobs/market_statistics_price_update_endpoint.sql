-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/market-statistics/aggregate endpoint
--
-- Purpose: Aggregates real-time market statistics by fetching current quotes for all stocks
--          from the screener table and categorizing them as up, down, or unchanged based on
--          their percent change. This provides real-time market sentiment data showing how
--          many stocks are advancing, declining, or unchanged. The aggregated statistics
--          are stored in Redis cache and later persisted to the database.
--
-- Schedule: Runs every 5 minutes, Monday to Friday, between 3PM-10PM WAT (14-21 UTC)
--           (*/5 14-21 * * 1-5)
--
-- Actual job

SELECT cron.schedule(
  'market-stats-aggregate-5m',
  '*/5 14-21 * * 1-5',  -- 2 PM - 9 PM UTC = 3 PM - 10 PM WAT
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/market-statistics/aggregate',
      body := '{}'::jsonb,
      headers := '{"Content-Type": "application/json"}'::jsonb,
      timeout_milliseconds := 600000
    );
  $$
);

-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/market-statistics/aggregate',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 600000
);
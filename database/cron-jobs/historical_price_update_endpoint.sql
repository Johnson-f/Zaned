-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/ingest/historicals?concurrency=8 endpoint
--
-- Purpose: Fetches historical price data for all stock symbols from the screener table.
--          Processes symbols concurrently (8 workers) to fetch:
--          - 10-year daily historical data (1d interval)
--          - 1-day 1-minute intraday data (aggregated to daily for screener price updates)
--          - 1-day 30-minute historical data for charting
--          All historical data is saved to Redis cache only (no immediate database writes).
--          The background persistence worker will later batch write this data to the database.
--
-- Schedule: Runs every 4 hours (0 */4 * * *)
--
-- Actual job

SELECT cron.schedule(
  'historicals-ingest-4h',
  '0 */4 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/ingest/historicals?concurrency=8',
      body := '{}'::jsonb,
      headers := '{"Content-Type": "application/json"}'::jsonb,
      timeout_milliseconds := 600000
    );
  $$
);

--- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/ingest/historicals?concurrency=8',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 600000
);
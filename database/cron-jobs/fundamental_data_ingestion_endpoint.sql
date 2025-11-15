-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/ingest/fundamental-data endpoint
--
-- Purpose: Fetches fundamental financial data for all stock symbols from the screener table.
--          This includes income statements, balance sheets, and cash flow statements for
--          both annual and quarterly frequencies. The data is saved to Redis cache only
--          (no immediate database writes). The background persistence worker will later
--          batch write this data to the database.
--
-- Schedule: Runs daily at 3:00 AM UTC (0 3 * * *)

-- Actual job
SELECT cron.schedule(
  'fundamental-data-ingestion-daily',
  '0 3 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/ingest/fundamental-data',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb,
      timeout_milliseconds := 1800000
    );
  $$
);

-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/ingest/fundamental-data',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 1800000
);


-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/ingest/company-data endpoint
--
-- Purpose: Fetches company information for all stock symbols from the screener table.
--          This includes detailed company data such as name, price, market cap, sector,
--          industry, employee count, and various return metrics. The data is saved to
--          Redis cache only (no immediate database writes). The background persistence
--          worker will later batch write this data to the database.
--
-- Schedule: Runs daily at 2:00 AM UTC (0 2 * * *)

-- Actual job
SELECT cron.schedule(
  'company-info-ingestion-daily',
  '0 2 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/ingest/company-data',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb,
      timeout_milliseconds := 600000
    );
  $$
);

-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/ingest/company-data',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 600000
);


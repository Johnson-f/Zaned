-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-ever endpoint
--
-- Purpose: Saves high volume ever screener results to the database. This identifies
--          stocks that have the highest trading volume of all time compared to their
--          historical average. This helps identify stocks experiencing unprecedented
--          trading activity, which may indicate major news or events.
--
-- Schedule: Runs daily at 8:50 PM UTC (20:50 UTC) - after market close
--           (50 20 * * *)

-- Actual job
SELECT cron.schedule(
  'save-high-volume-ever-daily',
  '50 20 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-ever',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb,
      timeout_milliseconds := 60000
    );
  $$
);

-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-ever',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 60000
);


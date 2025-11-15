-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-year endpoint
--
-- Purpose: Saves high volume year screener results to the database. This identifies
--          stocks that have the highest trading volume in the current year compared
--          to their historical average. This helps identify stocks with sustained
--          high trading activity over a longer period.
--
-- Schedule: Runs daily at 8:45 PM UTC (20:45 UTC) - after market close
--           (45 20 * * *)

-- Actual job
SELECT cron.schedule(
  'save-high-volume-year-daily',
  '45 20 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-year',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb,
      timeout_milliseconds := 60000
    );
  $$
);

-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-year',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 60000
);


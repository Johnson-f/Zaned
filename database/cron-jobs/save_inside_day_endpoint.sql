-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/screener/save-inside-day endpoint
--
-- Purpose: Saves inside day screener results to the database. An inside day occurs when
--          a stock's high and low are within the previous day's high and low range.
--          This cron job calculates and saves the symbols that meet this criteria,
--          which is useful for identifying potential breakout candidates.
--
-- Schedule: Runs daily at 8:35 PM UTC (20:35 UTC) - after market close
--           (35 20 * * *)

-- Actual job
SELECT cron.schedule(
  'save-inside-day-daily',
  '35 20 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/screener/save-inside-day',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb,
      timeout_milliseconds := 60000
    );
  $$
);

-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/screener/save-inside-day',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 60000
);


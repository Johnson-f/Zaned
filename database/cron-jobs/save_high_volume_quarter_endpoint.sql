-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-quarter endpoint
--
-- Purpose: Saves high volume quarter screener results to the database. This identifies
--          stocks that have the highest trading volume in the current quarter compared
--          to their historical average. High volume often indicates increased interest
--          or significant news events.
--
-- Schedule: Runs daily at 8:40 PM UTC (20:40 UTC) - after market close
--           (40 20 * * *)

-- Actual job
SELECT cron.schedule(
  'save-high-volume-quarter-daily',
  '40 20 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-quarter',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb,
      timeout_milliseconds := 60000
    );
  $$
);

-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/screener/save-high-volume-quarter',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 60000
);


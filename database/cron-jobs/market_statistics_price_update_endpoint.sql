-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/market-statistics/aggregate endpoint

-- Actually job
-- Cron job that calls the market statistics aggregate endpoint every 5 minutes (Mon-Fri, 3PM-10PM)

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
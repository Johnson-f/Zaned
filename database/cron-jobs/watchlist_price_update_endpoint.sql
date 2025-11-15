-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/watchlist/update-prices endpoint

-- Actually job
SELECT cron.schedule(
  'watchlist-price-update-15m',
  '*/15 15-22 * * 1-5',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/watchlist/update-prices',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb
    );
  $$
);


--- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/watchlist/update-prices',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 60000
);


-- Endpoints i'm doing
-- https://zaned-backennd.onrender.com/api/admin/ingest/historicals?concurrency=8 -> Update every 4 hours

-- https://zaned-backennd.onrender.com/api/admin/market-statistics/aggregate -> update every 5 minutes monday to friday between 3PMto 10PM


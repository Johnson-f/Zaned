-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/cache/persist endpoint
--
-- Purpose: Manually triggers the persistence worker to batch write all cached data from Redis
--          to the PostgreSQL database. This includes:
--          - Historical price data (10y/1d and 1d/30m intervals)
--          - Company info data
--          - Fundamental data (income statements, balance sheets, cash flow statements)
--          - Market statistics
--          All this data was previously saved to Redis only during ingestion to reduce
--          database writes. This cron job ensures the data is eventually persisted to the
--          database for long-term storage and querying.
--
-- Schedule:
--   - Weekdays (Mon-Fri): Every 6 hours at 00:00, 06:00, 12:00, and 18:00 UTC
--   - Saturday: Every 12 hours at 00:00 and 12:00 UTC
--
-- Weekday job: Every 6 hours Monday to Friday (at 00:00, 06:00, 12:00, 18:00)
SELECT cron.schedule(
  'cache-persist-weekday-6h',
  '0 0,6,12,18 * * 1-5',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/cache/persist',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb
    );
  $$
);

-- Saturday job: Every 12 hours on Saturday (at 00:00 and 12:00)
SELECT cron.schedule(
  'cache-persist-saturday-12h',
  '0 0,12 * * 6',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/cache/persist',
      headers := '{"Content-Type": "application/json"}'::jsonb,
      body := '{}'::jsonb
    );
  $$
);


-- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/cache/persist',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 60000
);


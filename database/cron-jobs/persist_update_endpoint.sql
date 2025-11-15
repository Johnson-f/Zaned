-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/cache/persist endpoint

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


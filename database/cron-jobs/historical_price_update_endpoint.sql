-- Cron job that calls the https://zaned-backennd.onrender.com/api/admin/ingest/historicals?concurrency=8 endpoint

-- Actual job

SELECT cron.schedule(
  'historicals-ingest-4h',
  '0 */4 * * *',
  $$
    SELECT net.http_post(
      url := 'https://zaned-backennd.onrender.com/api/admin/ingest/historicals?concurrency=8',
      body := '{}'::jsonb,
      headers := '{"Content-Type": "application/json"}'::jsonb,
      timeout_milliseconds := 600000
    );
  $$
);

--- Testing
SELECT net.http_post(
  url := 'https://zaned-backennd.onrender.com/api/admin/ingest/historicals?concurrency=8',
  body := '{}'::jsonb,
  headers := '{"Content-Type": "application/json"}'::jsonb,
  timeout_milliseconds := 600000
);
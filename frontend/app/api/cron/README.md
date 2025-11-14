# Vercel Cron Jobs

This directory contains API routes that are triggered by Vercel's cron job service to call backend endpoints at scheduled intervals.

## Setup Instructions

### 1. Environment Variables

Add the following environment variable to your Vercel project:

- `CRON_SECRET`: A secret string to protect your cron endpoints from unauthorized access
- `NEXT_PUBLIC_API_BASE_URL`: Your backend API URL (defaults to `https://zaned-backennd.onrender.com`)

### 2. Vercel Configuration

The `vercel.json` file in the root directory contains all cron job schedules. Vercel will automatically detect and configure these cron jobs when you deploy.

### 3. Cron Job Schedules

| Job | Schedule | Description |
|-----|----------|-------------|
| **Company Info Ingestion** | `0 2 * * *` | Daily at 2:00 AM UTC |
| **Historical Data Ingestion** | `0 */4 * * *` | Every 4 hours (00:00, 04:00, 08:00, 12:00, 16:00, 20:00 UTC) |
| **Fundamental Data Ingestion** | `0 3 * * 0` | Weekly on Sundays at 3:00 AM UTC |
| **Watchlist Price Update** | `*/15 * * * *` | Every 15 minutes |
| **Market Statistics Aggregation** | `*/5 * * * *` | Every 5 minutes |
| **Market Statistics EOD** | `30 20 * * *` | Daily at 8:30 PM UTC (4:30 PM ET) |
| **Save Inside Day** | `35 20 * * *` | Daily at 8:35 PM UTC (4:35 PM ET) |
| **Save High Volume Quarter** | `40 20 * * *` | Daily at 8:40 PM UTC (4:40 PM ET) |
| **Save High Volume Year** | `45 20 * * *` | Daily at 8:45 PM UTC (4:45 PM ET) |
| **Save High Volume Ever** | `50 20 * * *` | Daily at 8:50 PM UTC (4:50 PM ET) |

## Security

All cron endpoints are protected by:
1. Vercel's `x-vercel-cron` header (automatically set by Vercel)
2. Optional `CRON_SECRET` environment variable for additional security

## Testing

To test cron jobs locally, you can manually call the endpoints:

```bash
# Set CRON_SECRET in your .env.local
curl -H "Authorization: Bearer YOUR_CRON_SECRET" http://localhost:2000/api/cron/company-info
```

Or in development mode, if `CRON_SECRET` is not set, the endpoints will allow requests (for testing purposes only).

## Monitoring

Monitor your cron jobs in the Vercel dashboard under the "Cron Jobs" section. You can see execution logs, success/failure rates, and execution times.

## Troubleshooting

1. **Cron jobs not running**: Ensure `vercel.json` is in the root directory and properly formatted
2. **401 Unauthorized errors**: Check that `CRON_SECRET` is set in Vercel environment variables
3. **Backend errors**: Check the backend API is accessible and responding correctly
4. **Timezone issues**: All schedules are in UTC. Adjust accordingly for your timezone needs.


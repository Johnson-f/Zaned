# Quick Start: Deploy Backend with HTTPS

Get your backend running at `https://api.zaned.space` in 5 minutes.

## Prerequisites

- [x] Domain: `zaned.space` 
- [ ] Server with Docker installed
- [ ] DNS configured: `api.zaned.space` → Your server IP

## Step 1: Configure DNS

Add this A record to your domain:

```
Type: A
Name: api
Value: YOUR_SERVER_IP
TTL: 3600
```

Verify DNS propagation:
```bash
dig api.zaned.space
# or
nslookup api.zaned.space
```

## Step 2: Configure Environment

```bash
cd backend

# Copy environment template
cp .env.production .env

# Edit .env with your values
nano .env
```

Update these values in `.env`:
```bash
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_JWT_SECRET=your-jwt-secret
DATABASE_URL=postgresql://postgres:password@host:5432/postgres
ALLOWED_ORIGINS=https://zaned.space,https://www.zaned.space
```

## Step 3: Deploy

```bash
# Make deploy script executable
chmod +x deploy.sh

# Deploy!
./deploy.sh
```

That's it! Your API is now live at `https://api.zaned.space`

## Step 4: Test

```bash
# Test health endpoint
curl https://api.zaned.space/api/health

# Expected response:
# {"status":"ok","message":"Server is running"}
```

## Step 5: Update Frontend

In your frontend project:

```javascript
// .env.production
NEXT_PUBLIC_API_URL=https://api.zaned.site
```

---

## What Just Happened?

The deployment script:
1. ✓ Built your Docker image
2. ✓ Started backend service
3. ✓ Started Caddy reverse proxy
4. ✓ Obtained SSL certificate (automatic)
5. ✓ Configured HTTPS redirect
6. ✓ Set up security headers

## Useful Commands

```bash
# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Restart services
docker-compose -f docker-compose.prod.yml restart

# Stop services
docker-compose -f docker-compose.prod.yml down

# Check status
docker-compose -f docker-compose.prod.yml ps
```

## Troubleshooting

### DNS not working?
```bash
# Check DNS
dig api.zaned.site

# Wait for propagation (can take up to 48 hours)
```

### SSL certificate not obtained?
```bash
# Check Caddy logs
docker-compose -f docker-compose.prod.yml logs caddy

# Ensure ports 80 and 443 are open
sudo ufw allow 80
sudo ufw allow 443
```

### Backend not responding?
```bash
# Check backend logs
docker-compose -f docker-compose.prod.yml logs backend

# Check backend health directly
docker exec screener-backend wget -O- http://localhost:8080/api/health
```

### CORS errors from frontend?
```bash
# Update .env
ALLOWED_ORIGINS=https://zaned.site,https://www.zaned.site

# Restart backend
docker-compose -f docker-compose.prod.yml restart backend
```

---

## Next Steps

1. **Monitor your API**: Set up monitoring/alerts
2. **Backups**: Configure database backups
3. **Scaling**: Add more backend instances if needed
4. **CI/CD**: Automate deployments with GitHub Actions

## Documentation

- [HTTPS_SETUP.md](./HTTPS_SETUP.md) - Detailed HTTPS configuration
- [FRONTEND_INTEGRATION.md](./FRONTEND_INTEGRATION.md) - Frontend setup guide
- [DEPLOYMENT.md](./DEPLOYMENT.md) - General deployment guide

## Support

Having issues? Check:
1. Logs: `docker-compose -f docker-compose.prod.yml logs`
2. DNS: `dig api.zaned.site`
3. Firewall: Ports 80, 443 open?
4. SSL: Certificate obtained? Check Caddy logs

---

**Your API is ready! 🚀**

Access it at: `https://api.zaned.site/api/health`

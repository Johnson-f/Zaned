# HTTPS Setup Guide for zaned.site

This guide will help you configure HTTPS for your backend API at `api.zaned.site`.

## Prerequisites

1. Domain name: `zaned.site` (you have this ✓)
2. DNS A record pointing `api.zaned.site` to your server's IP address
3. Server with Docker and Docker Compose installed
4. Ports 80 and 443 open on your firewall

## Option 1: Caddy (Recommended - Easiest)

Caddy automatically obtains and renews SSL certificates. No manual certificate management needed!

### Setup Steps

1. **Update DNS**: Point `api.zaned.site` to your server IP
   ```
   A record: api.zaned.site → YOUR_SERVER_IP
   ```

2. **Update environment variables**:
   ```bash
   # Add to your .env file
   ALLOWED_ORIGINS=https://zaned.site,https://www.zaned.site
   ```

3. **Deploy with Caddy**:
   ```bash
   cd backend
   docker-compose -f docker-compose.prod.yml up -d
   ```

4. **That's it!** Caddy will automatically:
   - Obtain SSL certificates from Let's Encrypt
   - Renew certificates before expiry
   - Redirect HTTP to HTTPS
   - Configure secure SSL settings

### Verify Setup

```bash
# Check if services are running
docker-compose -f docker-compose.prod.yml ps

# Check Caddy logs
docker-compose -f docker-compose.prod.yml logs caddy

# Test the API
curl https://api.zaned.site/api/health
```

### Custom Domain Configuration

If you want to use a different subdomain, edit `Caddyfile`:

```caddyfile
your-subdomain.zaned.site {
    reverse_proxy backend:8080
    # ... rest of config
}
```

---

## Option 2: Nginx + Certbot (Traditional)

More control but requires manual certificate setup.

### Setup Steps

1. **Update DNS**: Point `api.zaned.site` to your server IP

2. **Update nginx.conf**: Replace `api.zaned.site` with your domain

3. **Obtain SSL certificate**:
   ```bash
   chmod +x setup-ssl.sh
   ./setup-ssl.sh api.zaned.site your-email@example.com
   ```

4. **Update docker-compose.prod.yml**:
   - Comment out the `caddy` service
   - Uncomment the `nginx` and `certbot` services

5. **Start services**:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Certificate Renewal

Certbot will automatically renew certificates. To manually renew:

```bash
docker-compose -f docker-compose.prod.yml run --rm certbot renew
docker-compose -f docker-compose.prod.yml restart nginx
```

---

## Frontend Configuration

Update your frontend to use the HTTPS API endpoint:

### React/Next.js

```javascript
// .env.production
NEXT_PUBLIC_API_URL=https://api.zaned.site

// or in your config
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'https://api.zaned.site';
```

### Vue/Nuxt

```javascript
// nuxt.config.js
export default {
  publicRuntimeConfig: {
    apiURL: process.env.API_URL || 'https://api.zaned.site'
  }
}
```

### Vanilla JavaScript

```javascript
const API_BASE_URL = 'https://api.zaned.site';

fetch(`${API_BASE_URL}/api/health`)
  .then(res => res.json())
  .then(data => console.log(data));
```

---

## DNS Configuration

Set up these DNS records for your domain:

```
Type    Name    Value               TTL
A       api     YOUR_SERVER_IP      3600
A       @       YOUR_FRONTEND_IP    3600
CNAME   www     zaned.site          3600
```

---

## Security Checklist

- [x] HTTPS enabled with valid SSL certificate
- [x] HTTP redirects to HTTPS
- [x] HSTS header enabled
- [x] Security headers configured
- [ ] Update CORS to allow only your frontend domain
- [ ] Set up firewall rules (allow only 80, 443, 22)
- [ ] Enable rate limiting (optional)
- [ ] Set up monitoring and alerts

---

## Troubleshooting

### Certificate Not Obtained

**Problem**: Caddy/Certbot can't obtain certificate

**Solutions**:
1. Verify DNS is pointing to your server: `dig api.zaned.site`
2. Check ports 80 and 443 are open: `netmap -tuln | grep -E '80|443'`
3. Ensure no other service is using ports 80/443
4. Check Caddy logs: `docker-compose -f docker-compose.prod.yml logs caddy`

### CORS Errors

**Problem**: Frontend can't access API due to CORS

**Solution**: Update `.env` file:
```bash
ALLOWED_ORIGINS=https://zaned.site,https://www.zaned.site
```

Then restart:
```bash
docker-compose -f docker-compose.prod.yml restart backend
```

### 502 Bad Gateway

**Problem**: Nginx/Caddy can't reach backend

**Solutions**:
1. Check backend is running: `docker-compose -f docker-compose.prod.yml ps`
2. Check backend health: `docker exec -it screener-backend wget -O- http://localhost:8080/api/health`
3. Check logs: `docker-compose -f docker-compose.prod.yml logs backend`

### Mixed Content Warnings

**Problem**: Frontend served over HTTPS but making HTTP requests

**Solution**: Ensure all API calls use `https://` not `http://`

---

## Production Deployment Checklist

Before going live:

- [ ] DNS records configured and propagated
- [ ] SSL certificate obtained and valid
- [ ] CORS configured with your frontend domain
- [ ] Environment variables set correctly
- [ ] Database connection working
- [ ] Health check endpoint responding
- [ ] Logs being collected
- [ ] Backups configured
- [ ] Monitoring set up

---

## Monitoring

### Check Service Status

```bash
# All services
docker-compose -f docker-compose.prod.yml ps

# Backend health
curl https://api.zaned.site/api/health

# SSL certificate expiry
echo | openssl s_client -servername api.zaned.site -connect api.zaned.site:443 2>/dev/null | openssl x509 -noout -dates
```

### View Logs

```bash
# All logs
docker-compose -f docker-compose.prod.yml logs -f

# Backend only
docker-compose -f docker-compose.prod.yml logs -f backend

# Caddy only
docker-compose -f docker-compose.prod.yml logs -f caddy
```

---

## Cost Considerations

- **SSL Certificates**: FREE (Let's Encrypt)
- **Caddy/Nginx**: FREE (open source)
- **Server**: Depends on your hosting provider
  - DigitalOcean: $6-12/month
  - AWS EC2: $5-20/month
  - Hetzner: €4-10/month

---

## Alternative: Cloud Platform with Built-in HTTPS

If you prefer not to manage SSL yourself, consider these platforms:

### Railway
```bash
# Automatic HTTPS, no configuration needed
railway up
```

### Render
- Automatic SSL certificates
- Custom domains with one click
- Free tier available

### Fly.io
```bash
# Automatic HTTPS
fly deploy
fly certs add api.zaned.site
```

These platforms handle SSL automatically but may cost more than self-hosting.

---

## Need Help?

Common issues and solutions:

1. **DNS not propagating**: Wait 24-48 hours or use `dig api.zaned.site` to check
2. **Port conflicts**: Stop other services using ports 80/443
3. **Certificate errors**: Ensure domain points to correct IP
4. **CORS issues**: Update ALLOWED_ORIGINS in .env

For more help, check the logs:
```bash
docker-compose -f docker-compose.prod.yml logs
```

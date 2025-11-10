# Docker Deployment Guide

## Prerequisites

- Docker installed
- Docker Compose installed (optional, for local testing)
- Environment variables configured

## Quick Start

### 1. Build the Docker Image

```bash
cd backend
docker build -t screener-backend:latest .
```

### 2. Run with Docker

```bash
docker run -d \
  --name screener-backend \
  -p 8080:8080 \
  -e SUPABASE_URL=your_supabase_url \
  -e SUPABASE_ANON_KEY=your_anon_key \
  -e SUPABASE_JWT_SECRET=your_jwt_secret \
  -e DATABASE_URL=your_database_url \
  -e PORT=8080 \
  screener-backend:latest
```

### 3. Run with Docker Compose (Recommended for Local Testing)

```bash
# Make sure .env file exists with your configuration
docker-compose up -d
```

## Environment Variables

Required environment variables:

- `SUPABASE_URL` - Your Supabase project URL
- `SUPABASE_ANON_KEY` - Your Supabase anonymous key
- `SUPABASE_JWT_SECRET` - Your Supabase JWT secret
- `DATABASE_URL` - PostgreSQL connection string
- `PORT` - Server port (default: 8080)

## Health Check

The application exposes a health check endpoint at:

```
GET /api/health
```

Response:
```json
{
  "status": "ok",
  "message": "Server is running"
}
```

## Production Deployment

### Docker Hub / Container Registry

1. **Tag your image:**
```bash
docker tag screener-backend:latest your-registry/screener-backend:v1.0.0
```

2. **Push to registry:**
```bash
docker push your-registry/screener-backend:v1.0.0
```

### Cloud Platforms

#### AWS ECS / Fargate

1. Push image to ECR
2. Create task definition with environment variables
3. Configure service with load balancer
4. Set health check path to `/api/health`

#### Google Cloud Run

```bash
gcloud run deploy screener-backend \
  --image your-registry/screener-backend:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars SUPABASE_URL=xxx,SUPABASE_ANON_KEY=xxx
```

#### Azure Container Instances

```bash
az container create \
  --resource-group myResourceGroup \
  --name screener-backend \
  --image your-registry/screener-backend:latest \
  --dns-name-label screener-backend \
  --ports 8080 \
  --environment-variables \
    SUPABASE_URL=xxx \
    SUPABASE_ANON_KEY=xxx
```

#### Railway / Render / Fly.io

These platforms can deploy directly from your Dockerfile:

1. Connect your Git repository
2. Set environment variables in the dashboard
3. Deploy automatically on push

## Monitoring

### Check Container Logs

```bash
docker logs screener-backend
```

### Check Container Health

```bash
docker inspect --format='{{.State.Health.Status}}' screener-backend
```

### Access Container Shell

```bash
docker exec -it screener-backend sh
```

## Troubleshooting

### Container Won't Start

1. Check logs: `docker logs screener-backend`
2. Verify environment variables are set
3. Ensure DATABASE_URL is accessible from container

### Database Connection Issues

- Verify DATABASE_URL format
- Check network connectivity
- Ensure SSL mode is configured correctly
- Verify database credentials

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Use different port
docker run -p 3000:8080 ...
```

## Optimization Tips

1. **Multi-stage builds** - Already implemented to reduce image size
2. **Layer caching** - Order Dockerfile commands from least to most frequently changing
3. **Health checks** - Already configured in docker-compose.yml
4. **Resource limits** - Add memory/CPU limits in production

Example with resource limits:
```bash
docker run -d \
  --name screener-backend \
  --memory="512m" \
  --cpus="1.0" \
  -p 8080:8080 \
  screener-backend:latest
```

## Security Best Practices

1. **Never commit .env files** - Already in .dockerignore
2. **Use secrets management** - AWS Secrets Manager, GCP Secret Manager, etc.
3. **Run as non-root user** - Consider adding USER directive in Dockerfile
4. **Scan images** - Use `docker scan screener-backend:latest`
5. **Keep base images updated** - Regularly rebuild with latest alpine

## Scaling

### Horizontal Scaling

Deploy multiple containers behind a load balancer:

```bash
docker-compose up -d --scale backend=3
```

### Kubernetes Deployment

Example deployment.yaml:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: screener-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: screener-backend
  template:
    metadata:
      labels:
        app: screener-backend
    spec:
      containers:
      - name: screener-backend
        image: your-registry/screener-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: SUPABASE_URL
          valueFrom:
            secretKeyRef:
              name: screener-secrets
              key: supabase-url
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

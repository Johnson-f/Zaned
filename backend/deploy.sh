#!/bin/bash

# Quick deployment script for production
set -e

echo "=== Deploying Screener Backend to Production ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    echo "Please create .env file from .env.production template"
    exit 1
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    exit 1
fi

# Check DNS configuration
echo -e "${YELLOW}Checking DNS configuration...${NC}"
if ! host api.zaned.site > /dev/null 2>&1; then
    echo -e "${YELLOW}Warning: DNS for api.zaned.site not configured or not propagated yet${NC}"
    echo "Make sure to point api.zaned.site to this server's IP address"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build the Docker image
echo -e "${GREEN}Building Docker image...${NC}"
docker-compose -f docker-compose.prod.yml build

# Stop existing containers
echo -e "${GREEN}Stopping existing containers...${NC}"
docker-compose -f docker-compose.prod.yml down

# Start services
echo -e "${GREEN}Starting services...${NC}"
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be healthy
echo -e "${GREEN}Waiting for services to start...${NC}"
sleep 10

# Check if backend is healthy
echo -e "${GREEN}Checking backend health...${NC}"
if docker-compose -f docker-compose.prod.yml exec -T backend wget -q -O- http://localhost:8080/api/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Backend is healthy${NC}"
else
    echo -e "${RED}✗ Backend health check failed${NC}"
    echo "Check logs with: docker-compose -f docker-compose.prod.yml logs backend"
    exit 1
fi

# Show running services
echo ""
echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
echo "Services running:"
docker-compose -f docker-compose.prod.yml ps
echo ""
echo "API URL: https://api.zaned.site"
echo "Health check: https://api.zaned.site/api/health"
echo ""
echo "View logs: docker-compose -f docker-compose.prod.yml logs -f"
echo "Stop services: docker-compose -f docker-compose.prod.yml down"
echo ""
echo -e "${YELLOW}Note: If using Caddy, SSL certificates will be obtained automatically.${NC}"
echo -e "${YELLOW}This may take a few minutes on first deployment.${NC}"

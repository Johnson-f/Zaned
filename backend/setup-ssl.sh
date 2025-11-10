#!/bin/bash

# SSL Setup Script for Nginx + Certbot
# This script helps you obtain SSL certificates using Let's Encrypt

set -e

echo "=== SSL Certificate Setup for api.zaned.site ==="
echo ""

# Check if domain is provided
DOMAIN=${1:-api.zaned.site}
EMAIL=${2:-admin@zaned.site}

echo "Domain: $DOMAIN"
echo "Email: $EMAIL"
echo ""

# Create directories
echo "Creating directories..."
mkdir -p certbot/conf
mkdir -p certbot/www

# Initial certificate request
echo "Requesting SSL certificate from Let's Encrypt..."
docker-compose -f docker-compose.prod.yml run --rm certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    -d $DOMAIN

echo ""
echo "✓ SSL certificate obtained successfully!"
echo ""
echo "Next steps:"
echo "1. Update nginx.conf with your domain name"
echo "2. Start the services: docker-compose -f docker-compose.prod.yml up -d"
echo "3. Certificates will auto-renew every 12 hours"

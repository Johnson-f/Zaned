.PHONY: help build run stop clean logs test docker-build docker-run docker-stop docker-clean

# Variables
IMAGE_NAME := screener-backend
CONTAINER_NAME := screener-backend
PORT := 8080

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go application locally
	go build -o main .

run: ## Run the application locally
	go run main.go

test: ## Run tests
	go test ./...

docker-build: ## Build Docker image
	docker build -t $(IMAGE_NAME):latest .

docker-run: ## Run Docker container
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p $(PORT):8080 \
		--env-file .env \
		$(IMAGE_NAME):latest

docker-stop: ## Stop Docker container
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true

docker-logs: ## View Docker container logs
	docker logs -f $(CONTAINER_NAME)

docker-shell: ## Access Docker container shell
	docker exec -it $(CONTAINER_NAME) sh

docker-clean: docker-stop ## Clean up Docker resources
	docker rmi $(IMAGE_NAME):latest || true

compose-up: ## Start services with docker-compose
	docker-compose up -d

compose-down: ## Stop services with docker-compose
	docker-compose down

compose-logs: ## View docker-compose logs
	docker-compose logs -f

prod-up: ## Start production services with HTTPS
	docker-compose -f docker-compose.prod.yml up -d

prod-down: ## Stop production services
	docker-compose -f docker-compose.prod.yml down

prod-logs: ## View production logs
	docker-compose -f docker-compose.prod.yml logs -f

prod-restart: ## Restart production services
	docker-compose -f docker-compose.prod.yml restart

ssl-setup: ## Setup SSL certificates (Nginx only)
	./setup-ssl.sh api.zaned.site

clean: ## Clean build artifacts
	rm -f main

dev: ## Run in development mode with hot reload (requires air)
	air

install-air: ## Install air for hot reload
	go install github.com/air-verse/air@latest

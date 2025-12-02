.PHONY: help build run test clean docker-build docker-push k8s-deploy k8s-delete

# Variables
SERVICE_NAME = order-service
IMAGE_NAME = your-registry/$(SERVICE_NAME)
VERSION ?= latest
PLATFORMS = linux/arm64,linux/amd64

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
install: ## Install dependencies
	go mod download
	go mod tidy

build: ## Build the binary
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/$(SERVICE_NAME) ./cmd/api

run: ## Run the service locally
	go run cmd/api/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

lint: ## Run linters
	golangci-lint run

fmt: ## Format code
	gofmt -s -w .
	go fmt ./...

vet: ## Run go vet
	go vet ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

# Docker
docker-build: ## Build Docker image for ARM64
	docker buildx build --platform linux/arm64 -t $(IMAGE_NAME):$(VERSION) .

docker-build-multi: ## Build Docker image for multiple platforms
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):$(VERSION) --push .

docker-push: ## Push Docker image
	docker push $(IMAGE_NAME):$(VERSION)

docker-run: ## Run Docker container locally
	docker run -p 8003:8003 --env-file .env $(IMAGE_NAME):$(VERSION)

# Docker Compose
compose-up: ## Start services with docker-compose
	docker-compose up -d

compose-down: ## Stop services with docker-compose
	docker-compose down

compose-logs: ## View docker-compose logs
	docker-compose logs -f

compose-ps: ## List docker-compose services
	docker-compose ps

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/postgres-pvc.yaml
	kubectl apply -f k8s/postgres-deployment.yaml
	kubectl apply -f k8s/redis-deployment.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	kubectl apply -f k8s/hpa.yaml
	kubectl apply -f k8s/networkpolicy.yaml

k8s-delete: ## Delete from Kubernetes
	kubectl delete -f k8s/networkpolicy.yaml
	kubectl delete -f k8s/hpa.yaml
	kubectl delete -f k8s/service.yaml
	kubectl delete -f k8s/deployment.yaml
	kubectl delete -f k8s/redis-deployment.yaml
	kubectl delete -f k8s/postgres-deployment.yaml
	kubectl delete -f k8s/postgres-pvc.yaml
	kubectl delete -f k8s/configmap.yaml
	kubectl delete -f k8s/secret.yaml

k8s-status: ## Check Kubernetes deployment status
	kubectl get pods -l app=$(SERVICE_NAME)
	kubectl get svc $(SERVICE_NAME)

k8s-logs: ## View Kubernetes logs
	kubectl logs -f deployment/$(SERVICE_NAME)

k8s-describe: ## Describe Kubernetes deployment
	kubectl describe deployment $(SERVICE_NAME)

k8s-restart: ## Restart Kubernetes deployment
	kubectl rollout restart deployment/$(SERVICE_NAME)

k8s-scale: ## Scale deployment (usage: make k8s-scale REPLICAS=3)
	kubectl scale deployment $(SERVICE_NAME) --replicas=$(REPLICAS)

# Database
db-migrate: ## Run database migrations
	psql $(DATABASE_URL) -f migrations/001_initial.sql

db-shell: ## Open PostgreSQL shell
	psql $(DATABASE_URL)

# Load Testing
load-test: ## Run load tests with k6
	k6 run tests/load/order-test.js

# CI/CD
ci: lint vet test ## Run CI checks

all: clean install build test ## Build everything

.DEFAULT_GOAL := help
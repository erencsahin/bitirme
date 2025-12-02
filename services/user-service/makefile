.PHONY: help install build dev docker-build docker-up docker-down docker-logs test clean deploy k8s-deploy k8s-delete k8s-logs k8s-status

# Variables
IMAGE_NAME := user-service
IMAGE_TAG := latest
REGISTRY := your-registry  # UPDATE THIS (e.g., ghcr.io/yourusername)
FULL_IMAGE := $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	npm install

build: ## Build TypeScript
	npm run build

dev: ## Run in development mode
	npm run dev

prisma-generate: ## Generate Prisma Client
	npx prisma generate

prisma-migrate: ## Run database migrations
	npx prisma migrate dev

prisma-studio: ## Open Prisma Studio
	npx prisma studio

test: ## Run tests
	npm test

clean: ## Clean build artifacts
	rm -rf dist node_modules

# Docker commands
docker-build: ## Build Docker image
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-build-push: ## Build and push Docker image to registry
	docker build --platform linux/amd64,linux/arm64 -t $(FULL_IMAGE) --push .

docker-up: ## Start services with docker-compose
	docker-compose up -d

docker-down: ## Stop services with docker-compose
	docker-compose down

docker-logs: ## View docker-compose logs
	docker-compose logs -f user-service

docker-ps: ## List running containers
	docker-compose ps

docker-shell: ## Shell into user-service container
	docker exec -it user-service sh

docker-migrate: ## Run migrations in Docker
	docker exec -it user-service npx prisma migrate deploy

# Kubernetes commands
k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/postgres-pvc.yaml
	kubectl apply -f k8s/postgres-deployment.yaml
	kubectl apply -f k8s/redis-deployment.yaml
	@echo "Waiting for databases to be ready..."
	kubectl wait --for=condition=ready pod -l app=postgres-user --timeout=120s
	kubectl wait --for=condition=ready pod -l app=redis-user --timeout=120s
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	kubectl apply -f k8s/hpa.yaml
	kubectl apply -f k8s/networkpolicy.yaml
	@echo "Deployment complete!"

k8s-delete: ## Delete from Kubernetes
	kubectl delete -f k8s/ --ignore-not-found=true

k8s-restart: ## Restart user-service pods
	kubectl rollout restart deployment/user-service

k8s-logs: ## View Kubernetes logs
	kubectl logs -f -l app=user-service --all-containers=true

k8s-status: ## Check deployment status
	kubectl get pods,svc,deploy -l app=user-service
	kubectl get pods,svc,deploy -l app=postgres-user
	kubectl get pods,svc,deploy -l app=redis-user

k8s-describe: ## Describe user-service deployment
	kubectl describe deployment user-service

k8s-shell: ## Shell into user-service pod
	kubectl exec -it $$(kubectl get pod -l app=user-service -o jsonpath='{.items[0].metadata.name}') -- sh

k8s-migrate: ## Run migrations in Kubernetes
	kubectl exec $$(kubectl get pod -l app=user-service -o jsonpath='{.items[0].metadata.name}') -- npx prisma migrate deploy

k8s-port-forward: ## Port forward to local
	kubectl port-forward svc/user-service 8001:8001

# Combined commands
deploy: docker-build-push k8s-deploy ## Build, push, and deploy to K8s

redeploy: docker-build-push k8s-restart ## Rebuild image and restart pods

local: docker-up ## Start local development environment

local-down: docker-down ## Stop local development environment
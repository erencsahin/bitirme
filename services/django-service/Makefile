.PHONY: help install migrate run test clean docker-build docker-push k8s-deploy

# Variables
SERVICE_NAME = product-service
IMAGE_NAME = your-registry/$(SERVICE_NAME)
VERSION ?= latest
PLATFORMS = linux/arm64,linux/amd64

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
install: ## Install Python dependencies
	pip install -r requirements.txt

migrate: ## Run database migrations
	python manage.py makemigrations
	python manage.py migrate

collectstatic: ## Collect static files
	python manage.py collectstatic --noinput

createsuperuser: ## Create Django superuser
	python manage.py createsuperuser

run: ## Run development server
	python manage.py runserver 0.0.0.0:8002

shell: ## Open Django shell
	python manage.py shell

test: ## Run tests
	python manage.py test

test-coverage: ## Run tests with coverage
	coverage run --source='.' manage.py test
	coverage report
	coverage html

lint: ## Run linters
	flake8 .
	black --check .
	isort --check-only .

format: ## Format code
	black .
	isort .

clean: ## Clean Python cache files
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	find . -type f -name "*.pyo" -delete
	find . -type f -name "*.coverage" -delete
	rm -rf htmlcov/
	rm -rf .pytest_cache/
	rm -rf .coverage

# Docker
docker-build: ## Build Docker image for ARM64
	docker buildx build --platform linux/arm64 -t $(IMAGE_NAME):$(VERSION) .

docker-build-multi: ## Build Docker image for multiple platforms
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):$(VERSION) --push .

docker-push: ## Push Docker image
	docker push $(IMAGE_NAME):$(VERSION)

docker-run: ## Run Docker container locally
	docker run -p 8002:8002 --env-file .env $(IMAGE_NAME):$(VERSION)

# Docker Compose
compose-up: ## Start services with docker-compose
	docker-compose up -d

compose-down: ## Stop services with docker-compose
	docker-compose down

compose-logs: ## View docker-compose logs
	docker-compose logs -f

compose-ps: ## List docker-compose services
	docker-compose ps

compose-rebuild: ## Rebuild and restart services
	docker-compose up -d --build

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -f k8s/django-secrets.yaml
	kubectl apply -f k8s/django-configmap.yaml
	kubectl apply -f k8s/postgres-pvc.yaml
	kubectl apply -f k8s/postgres-deployment.yaml
	kubectl apply -f k8s/redis-deployment.yaml
	kubectl apply -f k8s/django-deployment.yaml
	kubectl apply -f k8s/django-service.yaml
	kubectl apply -f k8s/hpa.yaml
	kubectl apply -f k8s/networkpolicy.yaml

k8s-delete: ## Delete from Kubernetes
	kubectl delete -f k8s/networkpolicy.yaml
	kubectl delete -f k8s/hpa.yaml
	kubectl delete -f k8s/django-service.yaml
	kubectl delete -f k8s/django-deployment.yaml
	kubectl delete -f k8s/redis-deployment.yaml
	kubectl delete -f k8s/postgres-deployment.yaml
	kubectl delete -f k8s/postgres-pvc.yaml
	kubectl delete -f k8s/django-configmap.yaml
	kubectl delete -f k8s/django-secrets.yaml

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
db-shell: ## Open PostgreSQL shell
	python manage.py dbshell

db-backup: ## Backup database
	python manage.py dumpdata > backup.json

db-restore: ## Restore database from backup
	python manage.py loaddata backup.json

# CI/CD
ci: lint test ## Run CI checks

all: clean install migrate collectstatic ## Setup everything

.DEFAULT_GOAL := help
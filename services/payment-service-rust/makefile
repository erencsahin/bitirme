.PHONY: help build run test docker-build docker-run k8s-deploy

help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run tests"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run with Docker Compose"
	@echo "  make k8s-deploy    - Deploy to Kubernetes"

build:
	cargo build --release

run:
	cargo run

test:
	cargo test

docker-build:
	docker build -t payment-service:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

k8s-deploy:
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/postgres-deployment.yaml
	kubectl apply -f k8s/postgres-pvc.yaml
	kubectl apply -f k8s/redis-deployment.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	kubectl apply -f k8s/hpa.yaml
	kubectl apply -f k8s/networkpolicy.yaml

k8s-delete:
	kubectl delete -f k8s/
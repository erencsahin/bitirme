#!/bin/bash

echo "🚀 Microservices K3s Deployment Başlıyor..."

# 1. Namespace
echo "📦 Namespace oluşturuluyor..."
kubectl apply -f namespace.yaml

# 2. Secrets
echo "🔐 Secrets oluşturuluyor..."
kubectl apply -f secrets.yaml

# 3. PostgreSQL
echo "🗄️  PostgreSQL StatefulSets deploy ediliyor..."
kubectl apply -f postgres-statefulset.yaml

# 4. Redis
echo "🔴 Redis Deployments deploy ediliyor..."
kubectl apply -f redis-deployments.yaml

# Veritabanlarının hazır olmasını bekle
echo "⏳ Veritabanları başlatılıyor (30 saniye)..."
sleep 30

# 5. OpenSearch Stack
echo "🔭 OpenSearch Stack deploy ediliyor..."
kubectl apply -f opensearch-stack-deployment.yaml

# OpenSearch'ün hazır olmasını bekle
echo "⏳ OpenSearch başlatılıyor (45 saniye)..."
sleep 45

# 6. Microservices
echo "🟢 User Service deploy ediliyor..."
kubectl apply -f user-service-deployment.yaml

echo "🟡 Product Service deploy ediliyor..."
kubectl apply -f product-service-deployment.yaml

echo "🔵 Order Service deploy ediliyor..."
kubectl apply -f order-service-deployment.yaml

echo "🟠 Payment Service deploy ediliyor..."
kubectl apply -f payment-service-deployment.yaml

echo ""
echo "✅ Deployment tamamlandı!"
echo ""
echo "📊 Durumu kontrol et:"
echo "   kubectl get pods -n microservices"
echo "   kubectl get svc -n microservices"
echo ""
echo "🔭 OpenSearch Dashboards:"
echo "   http://<raspberry-pi-ip>:30601"
echo ""
echo "🌐 Servis Endpoint'leri:"
echo "   User Service:    http://<raspberry-pi-ip>:30001/health"
echo "   Product Service: http://<raspberry-pi-ip>:30000/health/"
echo "   Order Service:   http://<raspberry-pi-ip>:30003/health"
echo "   Payment Service: http://<raspberry-pi-ip>:30085/api/health"

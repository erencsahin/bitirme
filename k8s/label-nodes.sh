#!/bin/bash

# Raspberry Pi node'larına label ekle
# Bu script'i K3s master node'unda çalıştır

echo "🏷️  Node'ları label'lıyoruz..."

# Node isimlerini al ve label ekle
kubectl label nodes pi-node-1 node-role=user-service --overwrite
kubectl label nodes pi-node-2 node-role=product-service --overwrite
kubectl label nodes pi-node-3 node-role=order-service --overwrite
kubectl label nodes pi-node-4 node-role=payment-service --overwrite

echo "✅ Node label'ları eklendi!"

# Kontrol et
kubectl get nodes --show-labels

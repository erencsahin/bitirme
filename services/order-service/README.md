# Order Service (Go)

Order management microservice built with Go, Gin, PostgreSQL, and OpenTelemetry.

## 🎯 Features

- ✅ Order creation with validation
- ✅ Order status management (pending, confirmed, shipped, delivered, cancelled)
- ✅ Order items with product integration
- ✅ User order history with pagination
- ✅ Redis caching for performance
- ✅ OpenTelemetry distributed tracing
- ✅ Health and readiness checks
- ✅ Graceful shutdown
- ✅ ARM64 support for Raspberry Pi

## 📋 Prerequisites

- Go 1.22+
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose
- Kubernetes (K3s) for production deployment

## 🚀 Quick Start

### Local Development

1. **Clone and setup:**
```bash
   git clone <repo>
   cd order-service
   cp .env.example .env
   # Edit .env with your configuration
```

2. **Install dependencies:**
```bash
   go mod download
   go mod tidy
```

3. **Run with Docker Compose:**
```bash
   docker-compose up -d
```

4. **Run locally:**
```bash
   go run cmd/api/main.go
```

5. **Test the service:**
```bash
   # Health check
   curl http://localhost:8003/health

   # Create order
   curl -X POST http://localhost:8003/api/orders \
     -H "Content-Type: application/json" \
     -d '{
       "user_id": "user123",
       "shipping_address": "123 Main St, City",
       "billing_address": "123 Main St, City",
       "order_items": [
         {
           "product_id": "prod123",
           "quantity": 2
         }
       ]
     }'
```

## 📡 API Endpoints

### Health Checks
- `GET /health` - Service health check
- `GET /ready` - Readiness check (includes DB)

### Orders
- `POST /api/orders` - Create new order
- `GET /api/orders` - List all orders (paginated)
- `GET /api/orders/:id` - Get order by ID
- `PATCH /api/orders/:id/status` - Update order status
- `POST /api/orders/:id/cancel` - Cancel order

### User Orders
- `GET /api/users/:user_id/orders` - Get user's orders (paginated)

### Request/Response Examples

**Create Order:**
```json
// Request
POST /api/orders
{
  "user_id": "user123",
  "shipping_address": "123 Main St, New York, NY 10001",
  "billing_address": "123 Main St, New York, NY 10001",
  "notes": "Please ring doorbell",
  "order_items": [
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "quantity": 2
    }
  ]
}

// Response
{
  "status": "success",
  "message": "Order created successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "user123",
    "status": "pending",
    "total_amount": 59.98,
    "currency": "USD",
    "shipping_address": "123 Main St, New York, NY 10001",
    "order_items": [
      {
        "id": "789e0123-e89b-12d3-a456-426614174000",
        "product_id": "550e8400-e29b-41d4-a716-446655440000",
        "quantity": 2,
        "unit_price": 29.99,
        "subtotal": 59.98
      }
    ],
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Update Order Status:**
```json
// Request
PATCH /api/orders/:id/status
{
  "status": "confirmed"
}

// Response
{
  "status": "success",
  "message": "Order status updated successfully"
}
```

**Pagination:**
```bash
GET /api/orders?page=1&page_size=10

Response:
{
  "status": "success",
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 45,
    "total_pages": 5
  }
}
```

## 🏗️ Architecture
```
┌─────────────────────────────────────────────┐
│           Order Service (Go)                 │
├─────────────────────────────────────────────┤
│                                              │
│  ┌──────────────┐      ┌──────────────┐    │
│  │   Handlers   │◄─────┤  Middleware  │    │
│  └──────┬───────┘      └──────────────┘    │
│         │                                    │
│  ┌──────▼───────┐                           │
│  │   Services   │                           │
│  └──────┬───────┘                           │
│         │                                    │
│  ┌──────▼───────┐      ┌──────────────┐    │
│  │ Repositories │◄─────┤    Cache     │    │
│  └──────┬───────┘      └──────────────┘    │
│         │                                    │
│  ┌──────▼───────┐                           │
│  │   Database   │                           │
│  └──────────────┘                           │
│                                              │
└─────────────────────────────────────────────┘
         │                    │
         │                    │
    ┌────▼────┐         ┌────▼────┐
    │  PGSQL  │         │  Redis  │
    └─────────┘         └─────────┘
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8003` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://...` |
| `REDIS_URL` | Redis connection string | `localhost:6379` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint | `localhost:4317` |
| `SERVICE_NAME` | Service name for tracing | `order-service` |
| `PRODUCT_SERVICE_URL` | Product service URL | `http://product-service:8002` |

## 🐳 Docker

### Build for ARM64 (Raspberry Pi):
```bash
docker buildx build --platform linux/arm64 -t order-service:latest .
```

### Build for multiple platforms:
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t order-service:latest .
```

### Push to registry:
```bash
docker tag order-service:latest your-registry/order-service:latest
docker push your-registry/order-service:latest
```

## ☸️ Kubernetes Deployment

### Deploy to K3s cluster:

1. **Create namespace (if needed):**
```bash
   kubectl create namespace microservices
```

2. **Apply configurations:**
```bash
   # Apply in order
   kubectl apply -f k8s/secret.yaml
   kubectl apply -f k8s/configmap.yaml
   kubectl apply -f k8s/postgres-pvc.yaml
   kubectl apply -f k8s/postgres-deployment.yaml
   kubectl apply -f k8s/redis-deployment.yaml
   kubectl apply -f k8s/deployment.yaml
   kubectl apply -f k8s/service.yaml
   kubectl apply -f k8s/hpa.yaml
   kubectl apply -f k8s/networkpolicy.yaml
```

3. **Verify deployment:**
```bash
   kubectl get pods -l app=order-service
   kubectl logs -f deployment/order-service
```

4. **Test service:**
```bash
   kubectl port-forward svc/order-service 8003:8003
   curl http://localhost:8003/health
```

## 🔍 Monitoring & Observability

### Logs
```bash
# View logs
kubectl logs -f deployment/order-service

# View logs with timestamps
kubectl logs -f deployment/order-service --timestamps

# View logs from specific pod
kubectl logs -f pod/order-service-xxxxx
```

### Metrics
- Prometheus scrapes metrics from `/metrics` endpoint
- HPA monitors CPU and memory usage

### Tracing
- OpenTelemetry exports traces to OTLP collector
- Distributed tracing across microservices
- View traces in Jaeger or similar

## 🧪 Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
go test ./... -tags=integration
```

### Load Testing with k6
```bash
k6 run tests/load/order-test.js
```

## 📊 Database Schema

### Orders Table
```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    shipping_address TEXT,
    billing_address TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

### Order Items Table
```sql
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

## 🔐 Security

- Non-root user in container
- Secret management via Kubernetes secrets
- Network policies for pod-to-pod communication
- HTTPS/TLS in production (via ingress)
- Input validation on all endpoints

## 📈 Performance

- Redis caching reduces database load
- Connection pooling for database
- HTTP client pooling for service-to-service calls
- Retry logic with exponential backoff
- Graceful degradation when dependencies fail

## 🐛 Troubleshooting

### Service won't start
```bash
# Check logs
kubectl logs deployment/order-service

# Check if database is ready
kubectl exec -it deployment/order-postgres -- psql -U orderuser -d orders -c "SELECT 1"

# Check if Redis is ready
kubectl exec -it deployment/order-redis -- redis-cli ping
```

### Cannot connect to Product Service
```bash
# Check if Product Service is running
kubectl get pods -l app=product-service

# Check service DNS
kubectl exec -it deployment/order-service -- nslookup product-service

# Check network policy
kubectl get networkpolicies
```

### High memory usage
```bash
# Check current resource usage
kubectl top pod -l app=order-service

# Scale down if needed
kubectl scale deployment order-service --replicas=1
```

## 📝 Development Guidelines

### Code Style
- Follow Go best practices and conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Add comments for exported functions

### Commit Messages
```
feat: add order cancellation endpoint
fix: correct order total calculation
docs: update API documentation
refactor: improve error handling in service layer
```

### Pull Requests
- Create feature branches from `main`
- Write descriptive PR descriptions
- Ensure all tests pass
- Update documentation if needed

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## 📄 License

MIT License

## 👥 Team

- **Developer**: Eren
- **Project**: Microservices Load Testing & Observability
- **Institution**: [Your University]

## 📚 References

- [Go Documentation](https://golang.org/doc/)
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
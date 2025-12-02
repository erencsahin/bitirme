# Payment Service - Rust

Payment processing microservice built with Rust and Axum.

## Features

- Payment processing (mock)
- Order integration
- JWT authentication
- Redis caching
- PostgreSQL database
- OpenTelemetry tracing
- Kubernetes ready

## Tech Stack

- Rust 1.75+
- Axum (web framework)
- SQLx (database)
- Redis
- PostgreSQL
- OpenTelemetry

## Getting Started

### Prerequisites

- Rust 1.75+
- PostgreSQL 15+
- Redis 7+

### Installation
```bash
# Build
cargo build --release

# Run
cargo run
```

### Docker
```bash
# Build and run
docker-compose up -d

# Stop
docker-compose down
```

### Kubernetes
```bash
# Deploy
make k8s-deploy

# Delete
make k8s-delete
```

## API Endpoints

- `GET /api/health` - Health check
- `POST /api/payments` - Create payment
- `GET /api/payments/:id` - Get payment by ID
- `GET /api/payments/order/:order_id` - Get payment by order ID

## Environment Variables
```env
PORT=8085
DATABASE_URL=postgres://postgres:postgres@localhost:5432/payment_db
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
ORDER_SERVICE_URL=http://localhost:8082
RUST_LOG=info
```
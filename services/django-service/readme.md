# Django Microservice - Products API

Product management microservice built with Django REST Framework.

## Features

- ✅ RESTful API for products and categories
- ✅ PostgreSQL database
- ✅ Redis caching
- ✅ OpenTelemetry instrumentation
- ✅ Standardized API responses
- ✅ Comprehensive filtering and search
- ✅ Pagination support
- ✅ Health check endpoint
- ✅ Docker support
- ✅ Kubernetes ready

## Tech Stack

- Python 3.11
- Django 5.0
- Django REST Framework 3.14
- PostgreSQL 15
- Redis 7
- OpenTelemetry
- Gunicorn

## Local Development

### Prerequisites

- Python 3.11+
- PostgreSQL 15+
- Redis 7+

### Setup

1. Create virtual environment:
```bash
python -m venv venv
source venv/bin/activate  # Linux/Mac
venv\Scripts\activate     # Windows
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

3. Create `.env` file (see `.env.example`)

4. Run migrations:
```bash
python manage.py migrate
```

5. Create superuser (optional):
```bash
python manage.py createsuperuser
```

6. Run development server:
```bash
python manage.py runserver
```

Server will run at: http://localhost:8000

## Docker Development

### Using Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f django

# Stop services
docker-compose down
```

### Build Docker Image
```bash
# For ARM64 (Raspberry Pi)
docker buildx build --platform linux/arm64 \
  -t 192.168.1.100:5000/django-service:v1 \
  --push .

# For AMD64 (x86_64)
docker buildx build --platform linux/amd64 \
  -t 192.168.1.100:5000/django-service:v1 \
  --push .
```

## API Endpoints

### Health & Metrics
- `GET /health/` - Health check
- `GET /metrics/` - Metrics

### Categories
- `GET /api/categories/` - List categories
- `GET /api/categories/{id}/` - Get category
- `POST /api/categories/` - Create category
- `PUT /api/categories/{id}/` - Update category
- `DELETE /api/categories/{id}/` - Delete category
- `GET /api/categories/{id}/products/` - Get category products

### Products
- `GET /api/products/` - List products
- `GET /api/products/{id}/` - Get product
- `POST /api/products/` - Create product
- `PUT /api/products/{id}/` - Update product
- `PATCH /api/products/{id}/` - Partial update
- `DELETE /api/products/{id}/` - Soft delete product
- `GET /api/products/statistics/` - Get statistics
- `GET /api/products/low_stock/` - Get low stock products
- `POST /api/products/{id}/update_stock/` - Update stock
- `GET /api/products/search/` - Search products

### Query Parameters

Products endpoint supports:
- `category` - Filter by category ID
- `in_stock` - Filter by stock (true/false)
- `is_active` - Filter by active status
- `min_price` - Minimum price
- `max_price` - Maximum price
- `search` - Search in name, description, SKU
- `ordering` - Sort by field
- `page` - Page number
- `page_size` - Items per page

## Response Format

### Success Response
```json
{
  "status": "success",
  "data": { ... },
  "timestamp": "2025-11-27T10:30:00Z"
}
```

### Error Response
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": { ... }
  },
  "timestamp": "2025-11-27T10:30:00Z"
}
```

### Paginated Response
```json
{
  "status": "success",
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  },
  "timestamp": "2025-11-27T10:30:00Z"
}
```

## Testing
```bash
# Run tests
python manage.py test

# Run with coverage
coverage run --source='.' manage.py test
coverage report
```

## Kubernetes Deployment
```bash
# Apply secrets
kubectl apply -f k8s/django-secrets.yaml

# Apply configmap
kubectl apply -f k8s/django-configmap.yaml

# Apply deployment
kubectl apply -f k8s/django-deployment.yaml

# Apply service
kubectl apply -f k8s/django-service.yaml

# Check status
kubectl get pods -l app=django-service
kubectl logs -f deployment/django-service
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SERVICE_NAME | Service name | django-service |
| SERVICE_VERSION | Service version | 1.0.0 |
| PORT | Server port | 8000 |
| DEBUG | Debug mode | False |
| SECRET_KEY | Django secret key | - |
| DB_HOST | PostgreSQL host | localhost |
| DB_PORT | PostgreSQL port | 5432 |
| DB_NAME | Database name | products_db |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | - |
| REDIS_HOST | Redis host | localhost |
| REDIS_PORT | Redis port | 6379 |
| OTEL_EXPORTER_OTLP_ENDPOINT | OpenTelemetry endpoint | http://localhost:4318 |

## Admin Panel

Access admin panel at: http://localhost:8000/admin/

Default credentials (after createsuperuser):
- Username: (your choice)
- Password: (your choice)

## License

MIT
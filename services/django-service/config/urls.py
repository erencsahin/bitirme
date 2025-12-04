# config/urls.py
from django.contrib import admin
from django.urls import path, include
from django.http import JsonResponse
from django.db import connection
from django.core.cache import cache
from drf_spectacular.views import (
    SpectacularAPIView,
    SpectacularSwaggerView,
    SpectacularRedocView,
)


# Health check views
def health_check(request):
    """Basic health check endpoint"""
    return JsonResponse({
        'status': 'healthy',
        'service': 'django-service',
        'version': '1.0.0',
    })


def readiness_check(request):
    """Readiness check with database and cache validation"""
    try:
        # Check database connection
        with connection.cursor() as cursor:
            cursor.execute("SELECT 1")
        db_status = 'ok'
    except Exception as e:
        db_status = 'error'
    
    try:
        # Check Redis/Cache connection
        cache.set('health_check_test', 'ok', 10)
        cache_result = cache.get('health_check_test')
        cache_status = 'ok' if cache_result == 'ok' else 'error'
    except Exception as e:
        cache_status = 'error'
    
    is_ready = db_status == 'ok' and cache_status == 'ok'
    status_code = 200 if is_ready else 503
    
    return JsonResponse({
        'status': 'ready' if is_ready else 'not_ready',
        'service': 'django-service',
        'version': '1.0.0',
        'checks': {
            'database': db_status,
            'cache': cache_status
        }
    }, status=status_code)


urlpatterns = [
    # Django Admin
    path('admin/', admin.site.urls),
    
    # Health Check Endpoints
    path('health/', health_check, name='health'),
    path('ready/', readiness_check, name='readiness'),
    
    # API Routes
    path('api/', include('products.urls')),
    
    # OpenAPI Schema (JSON)
    path('api/schema/', SpectacularAPIView.as_view(), name='schema'),
    
    # Swagger UI (Interactive API Documentation)
    path('api/docs/', SpectacularSwaggerView.as_view(url_name='schema'), name='swagger-ui'),
    
    # ReDoc (Alternative API Documentation)
    path('api/redoc/', SpectacularRedocView.as_view(url_name='schema'), name='redoc'),
]
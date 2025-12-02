# core/views.py
from django.http import JsonResponse
from django.conf import settings
from django.utils import timezone
from django.db import connection
from django.core.cache import cache
import logging

logger = logging.getLogger(__name__)


def health_check(request):
    """
    Health check endpoint
    Returns service health status with dependency checks
    """
    health_status = {
        'status': 'healthy',
        'service': settings.SERVICE_NAME,
        'version': settings.SERVICE_VERSION,
        'timestamp': timezone.now().isoformat(),
        'checks': {}
    }
    
    # Database check
    try:
        with connection.cursor() as cursor:
            cursor.execute("SELECT 1")
        health_status['checks']['database'] = 'ok'
    except Exception as e:
        health_status['checks']['database'] = 'error'
        health_status['status'] = 'unhealthy'
        logger.error(f"Database health check failed: {str(e)}")
    
    # Cache check
    try:
        cache.set('health_check', 'ok', 10)
        if cache.get('health_check') == 'ok':
            health_status['checks']['cache'] = 'ok'
        else:
            health_status['checks']['cache'] = 'error'
            health_status['status'] = 'degraded'
    except Exception as e:
        health_status['checks']['cache'] = 'error'
        health_status['status'] = 'degraded'
        logger.error(f"Cache health check failed: {str(e)}")
    
    status_code = 200 if health_status['status'] == 'healthy' else 503
    
    return JsonResponse(health_status, status=status_code)


def metrics(request):
    """
    Prometheus-style metrics endpoint
    """
    from products.models import Product, Category
    
    metrics_data = {
        'service': settings.SERVICE_NAME,
        'version': settings.SERVICE_VERSION,
        'timestamp': timezone.now().isoformat(),
        'metrics': {
            'products_total': Product.objects.filter(is_active=True).count(),
            'products_out_of_stock': Product.objects.filter(is_active=True, stock=0).count(),
            'categories_total': Category.objects.count(),
        }
    }
    
    return JsonResponse(metrics_data)
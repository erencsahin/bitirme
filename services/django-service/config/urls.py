# config/urls.py
from django.contrib import admin
from django.urls import path, include
from core.views import health_check, metrics

urlpatterns = [
    path('admin/', admin.site.urls),
    path('health/', health_check, name='health'),
    path('metrics/', metrics, name='metrics'),
    path('api/', include('products.urls')),
]
# products/urls.py
from django.urls import path, include
from rest_framework.routers import DefaultRouter
from .views import ProductViewSet, CategoryViewSet

router = DefaultRouter()
router.register(r'products', ProductViewSet, basename='product')
router.register(r'categories', CategoryViewSet, basename='category')

urlpatterns = [
    path('', include(router.urls)),
]

# Resulting Product URLs:
# GET    /api/products/                    - List products
# POST   /api/products/                    - Create product
# GET    /api/products/{id}/               - Get product
# PUT    /api/products/{id}/               - Update product
# DELETE /api/products/{id}/               - Delete product
# GET    /api/products/statistics/         - Product statistics
# GET    /api/products/low_stock/          - Low stock products
# POST   /api/products/{id}/update_stock/  - Update stock
# GET    /api/products/search/             - Search products

# Resulting Category URLs:
# GET    /api/categories/                  - List categories
# POST   /api/categories/                  - Create category
# GET    /api/categories/{id}/             - Get category
# PUT    /api/categories/{id}/             - Update category
# DELETE /api/categories/{id}/             - Delete category
# GET    /api/categories/{id}/products/    - Get category products

# Resulting Inventory URLs (NEW!):
# GET    /api/inventory/                          - List all inventories
# POST   /api/inventory/                          - Create inventory
# GET    /api/inventory/{product_id}/             - Get inventory by product
# PUT    /api/inventory/{product_id}/             - Update inventory
# DELETE /api/inventory/{product_id}/             - Delete inventory
# POST   /api/inventory/{product_id}/reserve/     - Reserve stock
# POST   /api/inventory/{product_id}/release/     - Release reserved stock
# POST   /api/inventory/{product_id}/increase/    - Increase stock
# POST   /api/inventory/{product_id}/decrease/    - Decrease stock
# GET    /api/inventory/{product_id}/check/?quantity=X  - Check availability
# GET    /api/inventory/low-stock/                - Get low stock items
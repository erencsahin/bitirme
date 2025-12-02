# products/views.py
from rest_framework import viewsets, status, filters
from rest_framework.decorators import action
from rest_framework.response import Response
from django.core.cache import cache
from django.db.models import Avg, Count
from django.utils.decorators import method_decorator
from django.views.decorators.cache import cache_page
import logging

from .models import Product, Category
from .serializers import (
    ProductListSerializer,
    ProductDetailSerializer,
    ProductCreateUpdateSerializer,
    CategorySerializer,
    StockCheckSerializer
)
from core.responses import StandardResponse
from core.pagination import StandardPagination

logger = logging.getLogger(__name__)


class CategoryViewSet(viewsets.ModelViewSet):
    """
    ViewSet for Category model
    
    Endpoints:
    - GET    /api/categories/          - List all categories
    - GET    /api/categories/{id}/     - Get single category
    - POST   /api/categories/          - Create category
    - PUT    /api/categories/{id}/     - Update category
    - DELETE /api/categories/{id}/     - Delete category
    - GET    /api/categories/{id}/products/ - Get category products
    """
    queryset = Category.objects.all()
    serializer_class = CategorySerializer
    pagination_class = StandardPagination
    
    @method_decorator(cache_page(60 * 15))
    def list(self, request, *args, **kwargs):
        logger.info("Fetching all categories")
        return super().list(request, *args, **kwargs)
    
    def retrieve(self, request, *args, **kwargs):
        instance = self.get_object()
        serializer = self.get_serializer(instance)
        return StandardResponse.success(data=serializer.data)
    
    def create(self, request, *args, **kwargs):
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        self.perform_create(serializer)
        
        logger.info(
            f"Category created: {serializer.data['name']}",
            extra={'category_id': serializer.data['id']}
        )
        
        return StandardResponse.success(
            data=serializer.data,
            status_code=status.HTTP_201_CREATED,
            message="Category created successfully"
        )
    
    def update(self, request, *args, **kwargs):
        partial = kwargs.pop('partial', False)
        instance = self.get_object()
        serializer = self.get_serializer(instance, data=request.data, partial=partial)
        serializer.is_valid(raise_exception=True)
        self.perform_update(serializer)
        
        logger.info(
            f"Category updated",
            extra={'category_id': instance.id}
        )
        
        return StandardResponse.success(
            data=serializer.data,
            message="Category updated successfully"
        )
    
    def destroy(self, request, *args, **kwargs):
        instance = self.get_object()
        instance.delete()
        
        logger.info(
            f"Category deleted",
            extra={'category_id': instance.id}
        )
        
        return StandardResponse.success(
            message="Category deleted successfully",
            status_code=status.HTTP_200_OK
        )
    
    @action(detail=True, methods=['get'])
    def products(self, request, pk=None):
        """Get all products in a category"""
        category = self.get_object()
        products = category.products.filter(is_active=True)
        
        page = self.paginate_queryset(products)
        if page is not None:
            serializer = ProductListSerializer(page, many=True)
            return self.get_paginated_response(serializer.data)
        
        serializer = ProductListSerializer(products, many=True)
        return StandardResponse.success(data=serializer.data)


class ProductViewSet(viewsets.ModelViewSet):
    """
    ViewSet for Product model with advanced features
    
    Endpoints:
    - GET    /api/products/                    - List products (with filters)
    - GET    /api/products/{id}/               - Get product details
    - POST   /api/products/                    - Create product
    - PUT    /api/products/{id}/               - Update product
    - PATCH  /api/products/{id}/               - Partial update
    - DELETE /api/products/{id}/               - Soft delete product
    - GET    /api/products/statistics/         - Get statistics
    - GET    /api/products/low_stock/          - Get low stock products
    - POST   /api/products/{id}/update_stock/  - Update stock
    - GET    /api/products/{id}/check_stock/   - Check stock availability
    - GET    /api/products/search/             - Search products
    """
    queryset = Product.objects.select_related('category').prefetch_related('images')
    pagination_class = StandardPagination
    filter_backends = [filters.SearchFilter, filters.OrderingFilter]
    search_fields = ['name', 'description', 'sku']
    ordering_fields = ['name', 'price', 'stock', 'rating', 'created_at']
    ordering = ['-created_at']
    
    def get_queryset(self):
        """Custom queryset with filtering"""
        queryset = super().get_queryset()
        
        # Filter by category
        category_id = self.request.query_params.get('category')
        if category_id:
            queryset = queryset.filter(category_id=category_id)
        
        # Filter by stock availability
        in_stock = self.request.query_params.get('in_stock')
        if in_stock is not None:
            if in_stock.lower() == 'true':
                queryset = queryset.filter(stock__gt=0)
            elif in_stock.lower() == 'false':
                queryset = queryset.filter(stock=0)
        
        # Filter by active status
        is_active = self.request.query_params.get('is_active')
        if is_active is not None:
            queryset = queryset.filter(is_active=is_active.lower() == 'true')
        
        # Price range filter
        min_price = self.request.query_params.get('min_price')
        max_price = self.request.query_params.get('max_price')
        if min_price:
            queryset = queryset.filter(price__gte=min_price)
        if max_price:
            queryset = queryset.filter(price__lte=max_price)
        
        return queryset
    
    def get_serializer_class(self):
        """Use different serializers for different actions"""
        if self.action == 'list':
            return ProductListSerializer
        elif self.action in ['create', 'update', 'partial_update']:
            return ProductCreateUpdateSerializer
        return ProductDetailSerializer
    
    @method_decorator(cache_page(60 * 5))
    def list(self, request, *args, **kwargs):
        """List products with caching"""
        logger.info(
            "Fetching products list",
            extra={'query_params': dict(request.query_params)}
        )
        return super().list(request, *args, **kwargs)
    
    def retrieve(self, request, *args, **kwargs):
        """Get single product with cache"""
        product_id = kwargs.get('pk')
        cache_key = f'product_detail_{product_id}'
        
        # Try to get from cache
        cached_data = cache.get(cache_key)
        if cached_data:
            logger.info(
                f"Product retrieved from cache",
                extra={'product_id': product_id}
            )
            return StandardResponse.success(data=cached_data)
        
        # If not in cache, get from DB
        instance = self.get_object()
        serializer = self.get_serializer(instance)
        
        # Cache the result for 10 minutes
        cache.set(cache_key, serializer.data, 60 * 10)
        logger.info(
            f"Product cached",
            extra={'product_id': product_id}
        )
        
        return StandardResponse.success(data=serializer.data)
    
    def create(self, request, *args, **kwargs):
        """Create new product"""
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        self.perform_create(serializer)
        
        logger.info(
            f"Product created: {serializer.data['name']}",
            extra={
                'product_id': serializer.data.get('id'),
                'sku': serializer.data['sku']
            }
        )
        
        # Get full product data
        product = Product.objects.get(id=serializer.data['id'])
        detail_serializer = ProductDetailSerializer(product)
        
        return StandardResponse.success(
            data=detail_serializer.data,
            status_code=status.HTTP_201_CREATED,
            message="Product created successfully"
        )
    
    def update(self, request, *args, **kwargs):
        """Update product and invalidate cache"""
        product_id = kwargs.get('pk')
        cache_key = f'product_detail_{product_id}'
        cache.delete(cache_key)
        
        partial = kwargs.pop('partial', False)
        instance = self.get_object()
        serializer = self.get_serializer(instance, data=request.data, partial=partial)
        serializer.is_valid(raise_exception=True)
        self.perform_update(serializer)
        
        logger.info(
            f"Product updated",
            extra={'product_id': product_id}
        )
        
        # Get full product data
        detail_serializer = ProductDetailSerializer(instance)
        
        return StandardResponse.success(
            data=detail_serializer.data,
            message="Product updated successfully"
        )
    
    def destroy(self, request, *args, **kwargs):
        """Soft delete - set is_active to False"""
        instance = self.get_object()
        instance.is_active = False
        instance.save()
        
        # Invalidate cache
        cache_key = f'product_detail_{instance.id}'
        cache.delete(cache_key)
        
        logger.info(
            f"Product soft deleted",
            extra={'product_id': instance.id}
        )
        
        return StandardResponse.success(
            message="Product deleted successfully",
            status_code=status.HTTP_200_OK
        )
    
    @action(detail=False, methods=['get'])
    def statistics(self, request):
        """Get product statistics"""
        stats = Product.objects.filter(is_active=True).aggregate(
            total_products=Count('id'),
            avg_price=Avg('price'),
            total_stock=Count('stock'),
            avg_rating=Avg('rating')
        )
        
        # Round decimal values
        if stats['avg_price']:
            stats['avg_price'] = float(round(stats['avg_price'], 2))
        if stats['avg_rating']:
            stats['avg_rating'] = float(round(stats['avg_rating'], 2))
        
        logger.info("Product statistics requested")
        
        return StandardResponse.success(data=stats)
    
    @action(detail=False, methods=['get'])
    def low_stock(self, request):
        """Get products with low stock"""
        threshold = int(request.query_params.get('threshold', 10))
        products = Product.objects.filter(
            is_active=True,
            stock__lt=threshold,
            stock__gt=0
        ).order_by('stock')
        
        page = self.paginate_queryset(products)
        if page is not None:
            serializer = ProductListSerializer(page, many=True)
            return self.get_paginated_response(serializer.data)
        
        serializer = ProductListSerializer(products, many=True)
        return StandardResponse.success(data=serializer.data)
    
    @action(detail=True, methods=['post'])
    def update_stock(self, request, pk=None):
        """Update product stock"""
        product = self.get_object()
        quantity = request.data.get('quantity')
        
        if quantity is None:
            return StandardResponse.error(
                code='MISSING_FIELD',
                message='quantity field is required',
                status_code=status.HTTP_400_BAD_REQUEST
            )
        
        try:
            quantity = int(quantity)
        except ValueError:
            return StandardResponse.error(
                code='INVALID_TYPE',
                message='quantity must be an integer',
                status_code=status.HTTP_400_BAD_REQUEST
            )
        
        new_stock = product.stock + quantity
        
        if new_stock < 0:
            return StandardResponse.error(
                code='INSUFFICIENT_STOCK',
                message=f'Insufficient stock. Available: {product.stock}',
                status_code=status.HTTP_400_BAD_REQUEST,
                details={'available_stock': product.stock, 'requested': abs(quantity)}
            )
        
        product.stock = new_stock
        product.save()
        
        # Invalidate cache
        cache_key = f'product_detail_{product.id}'
        cache.delete(cache_key)
        
        logger.info(
            f"Stock updated for product",
            extra={
                'product_id': product.id,
                'quantity_change': quantity,
                'new_stock': new_stock
            }
        )
        
        serializer = ProductDetailSerializer(product)
        return StandardResponse.success(
            data=serializer.data,
            message=f"Stock updated successfully. New stock: {new_stock}"
        )
    
    @action(detail=True, methods=['get'])
    def check_stock(self, request, pk=None):
        """
        Check if product has enough stock
        GET /api/products/{id}/check_stock/?quantity=5
        """
        product = self.get_object()
        
        quantity_param = request.query_params.get('quantity')
        if not quantity_param:
            return Response({
                'status': 'error',
                'message': 'quantity parameter is required'
            }, status=status.HTTP_400_BAD_REQUEST)
        
        try:
            quantity = int(quantity_param)
            if quantity < 1:
                raise ValueError("Quantity must be greater than 0")
        except ValueError:
            return Response({
                'status': 'error',
                'message': 'quantity must be a positive integer'
            }, status=status.HTTP_400_BAD_REQUEST)
        
        available = product.check_stock(quantity)
        
        logger.info(
            f"Stock check: Product {product.id}, Requested: {quantity}, Available: {available}",
            extra={'product_id': product.id, 'quantity': quantity}
        )
        
        return Response({
            'status': 'success',
            'data': {
                'product_id': product.id,
                'product_name': product.name,
                'requested_quantity': quantity,
                'available_stock': product.stock,
                'is_available': available
            }
        })
    
    @action(detail=False, methods=['get'])
    def search(self, request):
        """Search products (alias for queryset filtering)"""
        queryset = self.filter_queryset(self.get_queryset())
        
        page = self.paginate_queryset(queryset)
        if page is not None:
            serializer = self.get_serializer(page, many=True)
            return self.get_paginated_response(serializer.data)
        
        serializer = self.get_serializer(queryset, many=True)
        return StandardResponse.success(data=serializer.data)
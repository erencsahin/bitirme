# products/views.py
from rest_framework import viewsets, status, filters
from rest_framework.decorators import action
from django.core.cache import cache
from django.db.models import Avg, Count, F
from django.utils.decorators import method_decorator
from django.views.decorators.cache import cache_page
import logging

from .models import Product, Category, Inventory
from .serializers import (
    ProductListSerializer,
    ProductDetailSerializer,
    ProductCreateUpdateSerializer,
    CategorySerializer,
    InventorySerializer,
    InventoryCreateSerializer,
    InventoryUpdateSerializer,
    StockAdjustmentSerializer,
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


# ============================================
# INVENTORY VIEWSET - NEW!
# ============================================

class InventoryViewSet(viewsets.ModelViewSet):
    """
    ViewSet for Inventory management
    
    Endpoints:
    - GET    /api/inventory/                     - List all inventories
    - GET    /api/inventory/{product_id}/        - Get inventory by product
    - POST   /api/inventory/                     - Create inventory
    - PUT    /api/inventory/{product_id}/        - Update inventory
    - DELETE /api/inventory/{product_id}/        - Delete inventory
    - POST   /api/inventory/{product_id}/reserve/        - Reserve stock
    - POST   /api/inventory/{product_id}/release/        - Release reserved stock
    - POST   /api/inventory/{product_id}/increase/       - Increase stock
    - POST   /api/inventory/{product_id}/decrease/       - Decrease stock
    - GET    /api/inventory/{product_id}/check/          - Check availability
    - GET    /api/inventory/low-stock/                   - Get low stock items
    """
    
    queryset = Inventory.objects.select_related('product').all()
    pagination_class = StandardPagination
    lookup_field = 'product_id'
    lookup_url_kwarg = 'product_id'
    
    def get_serializer_class(self):
        """Use different serializers for different actions"""
        if self.action == 'create':
            return InventoryCreateSerializer
        elif self.action in ['update', 'partial_update']:
            return InventoryUpdateSerializer
        return InventorySerializer
    
    def list(self, request, *args, **kwargs):
        """List all inventories"""
        logger.info("Fetching all inventories")
        return super().list(request, *args, **kwargs)
    
    def retrieve(self, request, *args, **kwargs):
        """Get inventory by product ID"""
        try:
            inventory = self.get_object()
            serializer = self.get_serializer(inventory)
            return StandardResponse.success(data=serializer.data)
        except Inventory.DoesNotExist:
            return StandardResponse.error(
                code='NOT_FOUND',
                message=f'Inventory not found for product {kwargs.get("product_id")}',
                status_code=status.HTTP_404_NOT_FOUND
            )
    
    def create(self, request, *args, **kwargs):
        """Create inventory for a product"""
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        self.perform_create(serializer)
        
        # Get full inventory data
        inventory = Inventory.objects.get(product_id=serializer.instance.product_id)
        output_serializer = InventorySerializer(inventory)
        
        logger.info(
            f"Inventory created for product: {inventory.product_id}",
            extra={'product_id': inventory.product_id}
        )
        
        return StandardResponse.success(
            data=output_serializer.data,
            status_code=status.HTTP_201_CREATED,
            message="Inventory created successfully"
        )
    
    def update(self, request, *args, **kwargs):
        """Update inventory"""
        partial = kwargs.pop('partial', False)
        instance = self.get_object()
        serializer = self.get_serializer(instance, data=request.data, partial=partial)
        serializer.is_valid(raise_exception=True)
        self.perform_update(serializer)
        
        # Clear cache
        cache_key = f'inventory_product_{instance.product_id}'
        cache.delete(cache_key)
        
        output_serializer = InventorySerializer(instance)
        
        logger.info(
            f"Inventory updated for product: {instance.product_id}",
            extra={'product_id': instance.product_id}
        )
        
        return StandardResponse.success(
            data=output_serializer.data,
            message="Inventory updated successfully"
        )
    
    def destroy(self, request, *args, **kwargs):
        """Delete inventory"""
        instance = self.get_object()
        product_id = instance.product_id
        instance.delete()
        
        # Clear cache
        cache_key = f'inventory_product_{product_id}'
        cache.delete(cache_key)
        
        logger.info(
            f"Inventory deleted for product: {product_id}",
            extra={'product_id': product_id}
        )
        
        return StandardResponse.success(
            message="Inventory deleted successfully",
            status_code=status.HTTP_200_OK
        )
    
    @action(detail=True, methods=['post'])
    def reserve(self, request, product_id=None):
        """
        Reserve stock for an order
        POST /api/inventory/{product_id}/reserve/
        Body: {"quantity": 5, "reason": "Order #123"}
        """
        inventory = self.get_object()
        serializer = StockAdjustmentSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        
        quantity = serializer.validated_data['quantity']
        reason = serializer.validated_data.get('reason', '')
        
        try:
            inventory.reserve(quantity)
            
            # Clear cache
            cache_key = f'inventory_product_{product_id}'
            cache.delete(cache_key)
            
            logger.info(
                f"Stock reserved: Product {product_id}, Quantity: {quantity}, Reason: {reason}",
                extra={'product_id': product_id, 'quantity': quantity}
            )
            
            output_serializer = InventorySerializer(inventory)
            return StandardResponse.success(
                data=output_serializer.data,
                message=f"Successfully reserved {quantity} units"
            )
        
        except ValueError as e:
            return StandardResponse.error(
                code='INSUFFICIENT_STOCK',
                message=str(e),
                status_code=status.HTTP_400_BAD_REQUEST
            )
    
    @action(detail=True, methods=['post'])
    def release(self, request, product_id=None):
        """
        Release reserved stock
        POST /api/inventory/{product_id}/release/
        Body: {"quantity": 5, "reason": "Order cancelled"}
        """
        inventory = self.get_object()
        serializer = StockAdjustmentSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        
        quantity = serializer.validated_data['quantity']
        reason = serializer.validated_data.get('reason', '')
        
        try:
            inventory.release(quantity)
            
            # Clear cache
            cache_key = f'inventory_product_{product_id}'
            cache.delete(cache_key)
            
            logger.info(
                f"Stock released: Product {product_id}, Quantity: {quantity}, Reason: {reason}",
                extra={'product_id': product_id, 'quantity': quantity}
            )
            
            output_serializer = InventorySerializer(inventory)
            return StandardResponse.success(
                data=output_serializer.data,
                message=f"Successfully released {quantity} units"
            )
        
        except ValueError as e:
            return StandardResponse.error(
                code='INVALID_OPERATION',
                message=str(e),
                status_code=status.HTTP_400_BAD_REQUEST
            )
    
    @action(detail=True, methods=['post'])
    def increase(self, request, product_id=None):
        """
        Increase stock quantity
        POST /api/inventory/{product_id}/increase/
        Body: {"quantity": 100, "reason": "Restocking"}
        """
        inventory = self.get_object()
        serializer = StockAdjustmentSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        
        quantity = serializer.validated_data['quantity']
        reason = serializer.validated_data.get('reason', '')
        
        inventory.increase_stock(quantity)
        
        # Clear cache
        cache_key = f'inventory_product_{product_id}'
        cache.delete(cache_key)
        
        logger.info(
            f"Stock increased: Product {product_id}, Quantity: +{quantity}, Reason: {reason}",
            extra={'product_id': product_id, 'quantity': quantity}
        )
        
        output_serializer = InventorySerializer(inventory)
        return StandardResponse.success(
            data=output_serializer.data,
            message=f"Successfully increased stock by {quantity} units"
        )
    
    @action(detail=True, methods=['post'])
    def decrease(self, request, product_id=None):
        """
        Decrease stock quantity
        POST /api/inventory/{product_id}/decrease/
        Body: {"quantity": 10, "reason": "Damaged goods"}
        """
        inventory = self.get_object()
        serializer = StockAdjustmentSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        
        quantity = serializer.validated_data['quantity']
        reason = serializer.validated_data.get('reason', '')
        
        try:
            inventory.decrease_stock(quantity)
            
            # Clear cache
            cache_key = f'inventory_product_{product_id}'
            cache.delete(cache_key)
            
            logger.info(
                f"Stock decreased: Product {product_id}, Quantity: -{quantity}, Reason: {reason}",
                extra={'product_id': product_id, 'quantity': quantity}
            )
            
            output_serializer = InventorySerializer(inventory)
            return StandardResponse.success(
                data=output_serializer.data,
                message=f"Successfully decreased stock by {quantity} units"
            )
        
        except ValueError as e:
            return StandardResponse.error(
                code='INSUFFICIENT_STOCK',
                message=str(e),
                status_code=status.HTTP_400_BAD_REQUEST
            )
    
    @action(detail=True, methods=['get'])
    def check(self, request, product_id=None):
        """
        Check stock availability
        GET /api/inventory/{product_id}/check/?quantity=5
        """
        inventory = self.get_object()
        
        quantity_param = request.query_params.get('quantity')
        if not quantity_param:
            return StandardResponse.error(
                code='MISSING_PARAMETER',
                message='quantity parameter is required',
                status_code=status.HTTP_400_BAD_REQUEST
            )
        
        try:
            quantity = int(quantity_param)
            if quantity < 1:
                raise ValueError("Quantity must be greater than 0")
        except ValueError:
            return StandardResponse.error(
                code='INVALID_PARAMETER',
                message='quantity must be a positive integer',
                status_code=status.HTTP_400_BAD_REQUEST
            )
        
        available = inventory.check_availability(quantity)
        
        logger.info(
            f"Stock check: Product {product_id}, Requested: {quantity}, Available: {available}",
            extra={'product_id': product_id, 'quantity': quantity}
        )
        
        return StandardResponse.success(
            data={
                'available': available,
                'requested_quantity': quantity,
                'available_quantity': inventory.available_quantity,
                'total_quantity': inventory.quantity,
                'reserved_quantity': inventory.reserved_quantity
            }
        )
    
    @action(detail=False, methods=['get'])
    def low_stock(self, request):
        """
        Get products with low stock
        GET /api/inventory/low-stock/
        """
        low_stock_items = self.queryset.filter(
            quantity__lte=F('min_stock_level'),
            quantity__gt=0
        ).order_by('quantity')
        
        page = self.paginate_queryset(low_stock_items)
        if page is not None:
            serializer = self.get_serializer(page, many=True)
            return self.get_paginated_response(serializer.data)
        
        serializer = self.get_serializer(low_stock_items, many=True)
        
        logger.info(f"Low stock items retrieved: {low_stock_items.count()} items")
        
        return StandardResponse.success(data=serializer.data)
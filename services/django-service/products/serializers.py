# products/serializers.py
from rest_framework import serializers
from .models import Product, Category, ProductImage, Inventory


class ProductImageSerializer(serializers.ModelSerializer):
    class Meta:
        model = ProductImage
        fields = ['id', 'image_url', 'alt_text', 'is_primary']


class CategorySerializer(serializers.ModelSerializer):
    product_count = serializers.SerializerMethodField()
    
    class Meta:
        model = Category
        fields = ['id', 'name', 'description', 'product_count', 'created_at']
    
    def get_product_count(self, obj):
        return obj.products.filter(is_active=True).count()


# ============================================
# INVENTORY SERIALIZERS - NEW!
# ============================================

class InventorySerializer(serializers.ModelSerializer):
    """
    Full inventory serializer with all fields
    """
    product_id = serializers.IntegerField(source='product.id', read_only=True)
    product_name = serializers.CharField(source='product.name', read_only=True)
    product_sku = serializers.CharField(source='product.sku', read_only=True)
    available_quantity = serializers.IntegerField(read_only=True)
    status = serializers.CharField(read_only=True)
    
    class Meta:
        model = Inventory
        fields = [
            'product_id',
            'product_name',
            'product_sku',
            'quantity',
            'reserved_quantity',
            'available_quantity',
            'min_stock_level',
            'max_stock_level',
            'status',
            'warehouse_location',
            'created_at',
            'updated_at'
        ]
        read_only_fields = ['created_at', 'updated_at']


class InventoryCreateSerializer(serializers.ModelSerializer):
    """
    Serializer for creating inventory
    """
    product_id = serializers.IntegerField(write_only=True)
    
    class Meta:
        model = Inventory
        fields = [
            'product_id',
            'quantity',
            'min_stock_level',
            'max_stock_level',
            'warehouse_location'
        ]
    
    def validate_product_id(self, value):
        """Check if product exists"""
        try:
            Product.objects.get(id=value)
        except Product.DoesNotExist:
            raise serializers.ValidationError(f"Product with id {value} does not exist")
        
        # Check if inventory already exists
        if Inventory.objects.filter(product_id=value).exists():
            raise serializers.ValidationError(
                f"Inventory already exists for product {value}"
            )
        return value
    
    def create(self, validated_data):
        product_id = validated_data.pop('product_id')
        product = Product.objects.get(id=product_id)
        inventory = Inventory.objects.create(product=product, **validated_data)
        return inventory


class InventoryUpdateSerializer(serializers.ModelSerializer):
    """
    Serializer for updating inventory
    """
    class Meta:
        model = Inventory
        fields = [
            'quantity',
            'min_stock_level',
            'max_stock_level',
            'warehouse_location'
        ]


class StockAdjustmentSerializer(serializers.Serializer):
    """
    Serializer for stock adjustments (increase/decrease/reserve/release)
    """
    quantity = serializers.IntegerField(min_value=1)
    reason = serializers.CharField(
        max_length=200,
        required=False,
        allow_blank=True,
        help_text="Reason for stock adjustment"
    )


class StockCheckSerializer(serializers.Serializer):
    """
    Serializer for checking stock availability
    """
    quantity = serializers.IntegerField(min_value=1)


# ============================================
# PRODUCT SERIALIZERS (Updated with Inventory)
# ============================================

class ProductListSerializer(serializers.ModelSerializer):
    """Lightweight serializer for product lists"""
    category_name = serializers.CharField(source='category.name', read_only=True)
    primary_image = serializers.SerializerMethodField()
    
    class Meta:
        model = Product
        fields = [
            'id', 'name', 'sku', 'price', 'stock', 'in_stock',
            'category_name', 'primary_image', 'rating', 'review_count'
        ]
    
    def get_primary_image(self, obj):
        image = obj.images.filter(is_primary=True).first()
        if image:
            return image.image_url
        return None


class ProductDetailSerializer(serializers.ModelSerializer):
    """Detailed serializer for single product view - WITH INVENTORY"""
    category = CategorySerializer(read_only=True)
    category_id = serializers.IntegerField(write_only=True)
    images = ProductImageSerializer(many=True, read_only=True)
    inventory = InventorySerializer(read_only=True)  # ← NEW: Include inventory
    
    class Meta:
        model = Product
        fields = [
            'id', 'name', 'description', 'category', 'category_id',
            'price', 'stock', 'in_stock', 'sku', 'is_active',
            'rating', 'review_count', 'images', 'inventory',  # ← NEW: inventory field
            'created_at', 'updated_at'
        ]
        read_only_fields = ['created_at', 'updated_at']
    
    def validate_price(self, value):
        if value <= 0:
            raise serializers.ValidationError("Price must be greater than zero")
        return value
    
    def validate_stock(self, value):
        if value < 0:
            raise serializers.ValidationError("Stock cannot be negative")
        return value


class ProductCreateUpdateSerializer(serializers.ModelSerializer):
    """Serializer for creating/updating products"""
    
    class Meta:
        model = Product
        fields = [
            'name', 'description', 'category', 'price', 
            'stock', 'sku', 'is_active', 'rating'
        ]
    
    def validate_sku(self, value):
        instance = self.instance
        if instance and instance.sku != value:
            if Product.objects.filter(sku=value).exists():
                raise serializers.ValidationError("Product with this SKU already exists")
        elif not instance and Product.objects.filter(sku=value).exists():
            raise serializers.ValidationError("Product with this SKU already exists")
        return value
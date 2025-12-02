# products/serializers.py
from rest_framework import serializers
from .models import Product, Category, ProductImage


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
    """Detailed serializer for single product view"""
    category = CategorySerializer(read_only=True)
    category_id = serializers.IntegerField(write_only=True)
    images = ProductImageSerializer(many=True, read_only=True)
    
    class Meta:
        model = Product
        fields = [
            'id', 'name', 'description', 'category', 'category_id',
            'price', 'stock', 'in_stock', 'sku', 'is_active',
            'rating', 'review_count', 'images',
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


class StockCheckSerializer(serializers.Serializer):
    """Serializer for checking stock availability"""
    quantity = serializers.IntegerField(min_value=1)
# products/models.py
from django.db import models
from django.core.validators import MinValueValidator, MaxValueValidator
from decimal import Decimal


class Category(models.Model):
    """Product categories"""
    name = models.CharField(max_length=100, unique=True)
    description = models.TextField(blank=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    
    class Meta:
        verbose_name_plural = "Categories"
        ordering = ['name']
    
    def __str__(self):
        return self.name


class Product(models.Model):
    """Product model for inventory management"""
    
    name = models.CharField(max_length=200)
    description = models.TextField()
    category = models.ForeignKey(
        Category, 
        on_delete=models.CASCADE, 
        related_name='products'
    )
    price = models.DecimalField(
        max_digits=10, 
        decimal_places=2,
        validators=[MinValueValidator(Decimal('0.01'))]
    )
    stock = models.IntegerField(
        default=0,
        validators=[MinValueValidator(0)]
    )
    sku = models.CharField(max_length=50, unique=True)
    is_active = models.BooleanField(default=True)
    
    # Ratings
    rating = models.DecimalField(
        max_digits=3,
        decimal_places=2,
        default=0,
        validators=[
            MinValueValidator(0),
            MaxValueValidator(5)
        ]
    )
    review_count = models.IntegerField(default=0)
    
    # Timestamps
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    
    class Meta:
        ordering = ['-created_at']
        indexes = [
            models.Index(fields=['sku']),
            models.Index(fields=['category']),
            models.Index(fields=['-created_at']),
        ]
    
    def __str__(self):
        return f"{self.name} ({self.sku})"
    
    @property
    def in_stock(self):
        """Check if product is in stock"""
        return self.stock > 0


class ProductImage(models.Model):
    """Product images"""
    product = models.ForeignKey(
        Product,
        on_delete=models.CASCADE,
        related_name='images'
    )
    image_url = models.URLField(max_length=500)
    alt_text = models.CharField(max_length=200, blank=True)
    is_primary = models.BooleanField(default=False)
    created_at = models.DateTimeField(auto_now_add=True)
    
    class Meta:
        ordering = ['-is_primary', 'created_at']
    
    def __str__(self):
        return f"Image for {self.product.name}"


# ============================================
# INVENTORY MODEL - NEW!
# ============================================

class Inventory(models.Model):
    """
    Inventory management for products
    Tracks stock levels, reservations, and availability
    """
    
    product = models.OneToOneField(
        Product,
        on_delete=models.CASCADE,
        related_name='inventory',
        primary_key=True
    )
    
    quantity = models.IntegerField(
        default=0,
        validators=[MinValueValidator(0)],
        help_text="Total quantity in stock"
    )
    
    reserved_quantity = models.IntegerField(
        default=0,
        validators=[MinValueValidator(0)],
        help_text="Quantity reserved for pending orders"
    )
    
    min_stock_level = models.IntegerField(
        default=10,
        validators=[MinValueValidator(0)],
        help_text="Minimum stock level threshold"
    )
    
    max_stock_level = models.IntegerField(
        default=1000,
        validators=[MinValueValidator(0)],
        help_text="Maximum stock level threshold"
    )
    
    warehouse_location = models.CharField(
        max_length=100,
        blank=True,
        help_text="Physical warehouse location"
    )
    
    # Timestamps
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    
    class Meta:
        ordering = ['product']
        indexes = [
            models.Index(fields=['product']),
            models.Index(fields=['-updated_at']),
        ]
        verbose_name_plural = "Inventories"
    
    def __str__(self):
        return f"Inventory for {self.product.name} (SKU: {self.product.sku})"
    
    @property
    def available_quantity(self):
        """Calculate available quantity (total - reserved)"""
        return max(0, self.quantity - self.reserved_quantity)
    
    @property
    def status(self):
        """Calculate inventory status"""
        if self.available_quantity <= 0:
            return 'OUT_OF_STOCK'
        elif self.available_quantity <= self.min_stock_level:
            return 'LOW_STOCK'
        elif self.quantity >= self.max_stock_level:
            return 'OVERSTOCK'
        return 'IN_STOCK'
    
    def reserve(self, amount):
        """
        Reserve stock for an order
        Raises ValueError if insufficient stock
        """
        if self.available_quantity < amount:
            raise ValueError(
                f"Insufficient stock. Available: {self.available_quantity}, "
                f"Requested: {amount}"
            )
        self.reserved_quantity += amount
        self.save()
    
    def release(self, amount):
        """
        Release reserved stock (e.g., when order is cancelled)
        Raises ValueError if trying to release more than reserved
        """
        if self.reserved_quantity < amount:
            raise ValueError(
                f"Cannot release more than reserved. "
                f"Reserved: {self.reserved_quantity}, Requested: {amount}"
            )
        self.reserved_quantity -= amount
        self.save()
    
    def increase_stock(self, amount):
        """Increase total stock quantity"""
        self.quantity += amount
        self.save()
    
    def decrease_stock(self, amount):
        """
        Decrease total stock quantity
        Raises ValueError if insufficient available stock
        """
        if self.available_quantity < amount:
            raise ValueError(
                f"Insufficient available stock. Available: {self.available_quantity}, "
                f"Requested: {amount}"
            )
        self.quantity -= amount
        self.save()
    
    def check_availability(self, amount):
        """Check if requested amount is available"""
        return self.available_quantity >= amount
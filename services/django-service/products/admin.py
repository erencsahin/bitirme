from django.contrib import admin
from .models import Category, Product, ProductImage


@admin.register(Category)
class CategoryAdmin(admin.ModelAdmin):
    list_display = ['name', 'created_at', 'product_count']
    search_fields = ['name']
    ordering = ['name']
    
    def product_count(self, obj):
        return obj.products.count()
    product_count.short_description = 'Products'


class ProductImageInline(admin.TabularInline):
    model = ProductImage
    extra = 1
    fields = ['image_url', 'alt_text', 'is_primary']


@admin.register(Product)
class ProductAdmin(admin.ModelAdmin):
    list_display = ['name', 'sku', 'category', 'price', 'stock', 'is_active', 'rating']
    list_filter = ['category', 'is_active', 'created_at']
    search_fields = ['name', 'sku', 'description']
    ordering = ['-created_at']
    readonly_fields = ['created_at', 'updated_at']
    inlines = [ProductImageInline]
    
    fieldsets = (
        ('Basic Information', {
            'fields': ('name', 'sku', 'category', 'description')
        }),
        ('Pricing & Stock', {
            'fields': ('price', 'stock')
        }),
        ('Status & Rating', {
            'fields': ('is_active', 'rating', 'review_count')
        }),
        ('Timestamps', {
            'fields': ('created_at', 'updated_at'),
            'classes': ('collapse',)
        }),
    )
# core/pagination.py
from rest_framework.pagination import PageNumberPagination
from collections import OrderedDict


class StandardPagination(PageNumberPagination):
    """Standardized pagination"""
    page_size = 20
    page_size_query_param = 'page_size'
    max_page_size = 100
    
    def get_paginated_response(self, data):
        """Return standardized paginated response"""
        from .responses import StandardResponse
        
        pagination_data = OrderedDict([
            ('page', self.page.number),
            ('page_size', self.page.paginator.per_page),
            ('total', self.page.paginator.count),
            ('total_pages', self.page.paginator.num_pages),
        ])
        
        return StandardResponse.paginated(data, pagination_data)
# core/responses.py
from rest_framework.response import Response
from rest_framework import status
from django.utils import timezone


class StandardResponse:
    """Standardized API response format"""
    
    @staticmethod
    def success(data=None, status_code=status.HTTP_200_OK, message=None):
        """Success response"""
        response_data = {
            'status': 'success',
            'timestamp': timezone.now().isoformat(),
        }
        
        if data is not None:
            response_data['data'] = data
            
        if message:
            response_data['message'] = message
        
        return Response(response_data, status=status_code)
    
    @staticmethod
    def error(code, message, status_code=status.HTTP_400_BAD_REQUEST, details=None):
        """Error response"""
        response_data = {
            'status': 'error',
            'timestamp': timezone.now().isoformat(),
            'error': {
                'code': code,
                'message': message,
            }
        }
        
        if details:
            response_data['error']['details'] = details
        
        return Response(response_data, status=status_code)
    
    @staticmethod
    def paginated(items, pagination_data, status_code=status.HTTP_200_OK):
        """Paginated response"""
        response_data = {
            'status': 'success',
            'timestamp': timezone.now().isoformat(),
            'data': {
                'items': items,
                'pagination': pagination_data
            }
        }
        
        return Response(response_data, status=status_code)
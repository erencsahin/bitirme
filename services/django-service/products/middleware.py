# products/middleware.py
import logging
import requests
from django.http import JsonResponse
from django.conf import settings

logger = logging.getLogger(__name__)


class JWTAuthenticationMiddleware:
    """
    Middleware to authenticate requests using JWT tokens from User Service.
    Bypasses authentication for public endpoints.
    """
    
    # Public endpoints that don't require authentication
    PUBLIC_PATHS = [
        '/health/',
        '/ready/',
        '/admin/',
        '/api/schema/',
        '/api/docs/',
        '/api/redoc/',
        '/static/',
        '/api/products/',  # Allow public product listing
        '/api/categories/',  # Allow public category listing
    ]
    
    # User Service URL for token validation
    USER_SERVICE_URL = 'http://user-service:8001'  # Container name
    
    def __init__(self, get_response):
        self.get_response = get_response
    
    def __call__(self, request):
        path = request.path
        method = request.method
        
        # Bypass authentication for public paths
        if self._is_public_path(path):
            return self.get_response(request)
        
        # Bypass authentication for GET requests on product endpoints
        if method == 'GET' and (path.startswith('/api/products/') or path.startswith('/api/categories/')):
            return self.get_response(request)
        
        # For POST, PUT, PATCH, DELETE - require authentication
        if method in ['POST', 'PUT', 'PATCH', 'DELETE']:
            auth_header = request.META.get('HTTP_AUTHORIZATION', '')
            
            if not auth_header:
                return JsonResponse({
                    'status': 'error',
                    'message': 'Authorization header required'
                }, status=401)
            
            # Validate token with User Service
            is_valid, user_data = self._validate_token(auth_header)
            
            if not is_valid:
                return JsonResponse({
                    'status': 'error',
                    'message': 'Invalid or expired token'
                }, status=401)
            
            # Attach user data to request
            request.user_data = user_data
        
        response = self.get_response(request)
        return response
    
    def _is_public_path(self, path):
        """Check if the path is in public paths list"""
        for public_path in self.PUBLIC_PATHS:
            if path.startswith(public_path):
                return True
        return False
    
    def _validate_token(self, auth_header):
        """
        Validate JWT token with User Service
        Returns: (is_valid: bool, user_data: dict or None)
        """
        try:
            # Extract token from "Bearer <token>"
            if not auth_header.startswith('Bearer '):
                return False, None
            
            token = auth_header.replace('Bearer ', '').strip()
            
            # Call User Service to validate token
            response = requests.post(
                f'{self.USER_SERVICE_URL}/api/auth/validate',
                json={'token': token},
                timeout=5
            )
            
            if response.status_code == 200:
                data = response.json()
                if data.get('status') == 'success':
                    return True, data.get('data', {}).get('user', {})
            
            return False, None
            
        except requests.exceptions.RequestException as e:
            logger.error(f"Failed to validate token with User Service: {e}")
            # In development, allow requests if User Service is down
            if settings.DEBUG:
                logger.warning("Debug mode: Allowing request despite User Service error")
                return True, {'id': 'debug-user', 'email': 'debug@example.com'}
            return False, None
        except Exception as e:
            logger.error(f"Unexpected error in token validation: {e}")
            return False, None
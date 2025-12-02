import os
from django.http import JsonResponse
from .clients import UserServiceClient

user_service_client = UserServiceClient(
    base_url=os.getenv('USER_SERVICE_URL', 'http://localhost:8083')
)


class JWTAuthenticationMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        # Skip auth for health check
        if request.path in ['/health/', '/api/health/']:
            return self.get_response(request)

        # Get token from Authorization header
        auth_header = request.headers.get('Authorization', '')
        
        if not auth_header.startswith('Bearer '):
            return JsonResponse({
                'status': 'error',
                'message': 'Authorization header required'
            }, status=401)

        token = auth_header[7:]  # Remove 'Bearer '

        # Validate token with User Service
        is_valid = user_service_client.validate_token(token)
        
        if not is_valid:
            return JsonResponse({
                'status': 'error',
                'message': 'Invalid or expired token'
            }, status=401)

        # Get user ID and attach to request
        user_id = user_service_client.get_user_id_from_token(token)
        request.user_id = user_id

        response = self.get_response(request)
        return response
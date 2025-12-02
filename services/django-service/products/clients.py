import requests
import logging
from typing import Optional, Dict
from opentelemetry import trace

logger = logging.getLogger(__name__)
tracer = trace.get_tracer(__name__)


class UserServiceClient:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({'Content-Type': 'application/json'})

    def validate_token(self, token: str) -> bool:
        """Validate JWT token with User Service"""
        with tracer.start_as_current_span("user-service-validate-token") as span:
            try:
                url = f"{self.base_url}/api/auth/validate"
                headers = {'Authorization': f'Bearer {token}'}
                
                response = self.session.post(url, headers=headers, timeout=5)
                
                if response.status_code == 200:
                    data = response.json()
                    is_valid = data.get('status') == 'success'
                    span.set_attribute('token.valid', is_valid)
                    logger.info(f"Token validation result: {is_valid}")
                    return is_valid
                else:
                    logger.warning(f"Token validation failed: {response.status_code}")
                    return False
                    
            except requests.RequestException as e:
                logger.error(f"Error validating token: {e}")
                span.record_exception(e)
                return False

    def get_user_id_from_token(self, token: str) -> Optional[str]:
        """Get user ID from token"""
        with tracer.start_as_current_span("user-service-get-user-id") as span:
            try:
                url = f"{self.base_url}/api/auth/validate"
                headers = {'Authorization': f'Bearer {token}'}
                
                response = self.session.post(url, headers=headers, timeout=5)
                
                if response.status_code == 200:
                    data = response.json()
                    user_id = data.get('data', {}).get('userId')
                    span.set_attribute('user.id', user_id if user_id else 'none')
                    return user_id
                else:
                    return None
                    
            except requests.RequestException as e:
                logger.error(f"Error getting user ID: {e}")
                span.record_exception(e)
                return None
# core/exceptions.py
from rest_framework.views import exception_handler
from rest_framework import status
from .responses import StandardResponse
import logging

logger = logging.getLogger(__name__)


def custom_exception_handler(exc, context):
    """Custom exception handler for standardized error responses"""
    
    # Call REST framework's default exception handler first
    response = exception_handler(exc, context)
    
    if response is not None:
        # Log the error
        logger.error(
            f"API Error: {exc.__class__.__name__}",
            extra={
                'exception': str(exc),
                'status_code': response.status_code,
                'path': context['request'].path,
                'method': context['request'].method,
            }
        )
        
        # Standardize error response
        error_code = exc.__class__.__name__.upper().replace('EXCEPTION', '_ERROR')
        error_message = str(exc)
        
        # Handle validation errors
        if hasattr(exc, 'detail'):
            if isinstance(exc.detail, dict):
                details = exc.detail
            else:
                details = {'detail': exc.detail}
        else:
            details = None
        
        return StandardResponse.error(
            code=error_code,
            message=error_message,
            status_code=response.status_code,
            details=details
        )
    
    # If response is None, it's an unhandled exception
    logger.exception(
        f"Unhandled exception: {exc.__class__.__name__}",
        extra={
            'exception': str(exc),
            'path': context['request'].path,
            'method': context['request'].method,
        }
    )
    
    return StandardResponse.error(
        code='INTERNAL_SERVER_ERROR',
        message='An unexpected error occurred',
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR
    )
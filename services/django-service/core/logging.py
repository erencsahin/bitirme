# core/logging.py
import logging
import json
from django.utils import timezone
from django.conf import settings


class JSONFormatter(logging.Formatter):
    """JSON log formatter for structured logging"""
    
    def format(self, record):
        log_data = {
            'timestamp': timezone.now().isoformat(),
            'level': record.levelname,
            'service': settings.SERVICE_NAME,
            'message': record.getMessage(),
            'module': record.module,
            'function': record.funcName,
            'line': record.lineno,
        }
        
        # Add extra fields if present
        if hasattr(record, 'extra'):
            log_data['extra'] = record.extra
        
        # Add exception info if present
        if record.exc_info:
            log_data['exception'] = self.formatException(record.exc_info)
        
        # Add trace context if present (from OpenTelemetry)
        if hasattr(record, 'otelSpanID'):
            log_data['trace_id'] = record.otelTraceID
            log_data['span_id'] = record.otelSpanID
        
        return json.dumps(log_data)
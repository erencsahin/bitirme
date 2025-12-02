# config/tracing.py
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource, SERVICE_NAME, SERVICE_VERSION
from opentelemetry.instrumentation.django import DjangoInstrumentor
from opentelemetry.instrumentation.psycopg2 import Psycopg2Instrumentor
from opentelemetry.instrumentation.redis import RedisInstrumentor
from django.conf import settings
import logging

logger = logging.getLogger(__name__)


def setup_tracing():
    """Setup OpenTelemetry tracing"""
    try:
        # Create resource
        resource = Resource(attributes={
            SERVICE_NAME: settings.OTEL_SERVICE_NAME,
            SERVICE_VERSION: settings.SERVICE_VERSION,
        })
        
        # Create tracer provider
        provider = TracerProvider(resource=resource)
        
        # Create OTLP exporter
        otlp_exporter = OTLPSpanExporter(
            endpoint=f"{settings.OTEL_EXPORTER_OTLP_ENDPOINT}/v1/traces",
        )
        
        # Add span processor
        provider.add_span_processor(BatchSpanProcessor(otlp_exporter))
        
        # Set global tracer provider
        trace.set_tracer_provider(provider)
        
        # Instrument Django
        DjangoInstrumentor().instrument()
        
        # Instrument Psycopg2 (PostgreSQL)
        Psycopg2Instrumentor().instrument()
        
        # Instrument Redis
        RedisInstrumentor().instrument()
        
        logger.info("OpenTelemetry tracing initialized successfully")
        
    except Exception as e:
        logger.error(f"Failed to initialize OpenTelemetry: {str(e)}")
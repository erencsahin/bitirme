"""
OpenTelemetry configuration for Django service
"""
import os
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.django import DjangoInstrumentor
from opentelemetry.instrumentation.psycopg2 import Psycopg2Instrumentor
from opentelemetry.instrumentation.redis import RedisInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource


def configure_opentelemetry():
    """Configure OpenTelemetry for Django service"""
    
    # Service bilgileri
    service_name = os.getenv('OTEL_SERVICE_NAME', 'django-service')
    service_version = os.getenv('SERVICE_VERSION', '1.0.0')
    environment = os.getenv('ENVIRONMENT', 'development')
    
    # Resource oluştur
    resource = Resource.create({
        "service.name": service_name,
        "service.version": service_version,
        "deployment.environment": environment,
    })
    
    # TracerProvider'ı ayarla
    trace.set_tracer_provider(TracerProvider(resource=resource))
    tracer_provider = trace.get_tracer_provider()
    
    # OTLP Exporter
    otlp_endpoint = os.getenv('OTEL_EXPORTER_OTLP_ENDPOINT', 'http://localhost:4318/v1/traces')
    otlp_exporter = OTLPSpanExporter(endpoint=otlp_endpoint)
    
    # Span Processor
    span_processor = BatchSpanProcessor(otlp_exporter)
    tracer_provider.add_span_processor(span_processor)
    
    # Auto-instrumentation
    DjangoInstrumentor().instrument()
    Psycopg2Instrumentor().instrument()
    RedisInstrumentor().instrument()
    RequestsInstrumentor().instrument()
    
    print(f"✅ OpenTelemetry configured for {service_name}")
    print(f"📡 Sending traces to: {otlp_endpoint}")
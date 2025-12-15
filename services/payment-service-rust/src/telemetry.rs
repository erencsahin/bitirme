use opentelemetry::{global, KeyValue};
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::{
    runtime,
    trace::{self, RandomIdGenerator, Sampler},
    Resource,
};
use std::time::Duration;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

pub fn init_telemetry() -> anyhow::Result<()> {
    // Get configuration from environment
    let service_name = std::env::var("OTEL_SERVICE_NAME")
        .unwrap_or_else(|_| "payment-service".to_string());
    
    let service_version = std::env::var("SERVICE_VERSION")
        .unwrap_or_else(|_| "1.0.0".to_string());
    
    let environment = std::env::var("ENVIRONMENT")
        .unwrap_or_else(|_| "development".to_string());
    
    let otlp_endpoint = std::env::var("OTEL_ENDPOINT")
        .unwrap_or_else(|_| "http://localhost:4317".to_string());

    // Create OTLP exporter
    let tracer = opentelemetry_otlp::new_pipeline()
        .tracing()
        .with_exporter(
            opentelemetry_otlp::new_exporter()
                .tonic()
                .with_endpoint(otlp_endpoint.clone())
                .with_timeout(Duration::from_secs(3)),
        )
        .with_trace_config(
            trace::config()
                .with_sampler(Sampler::AlwaysOn)
                .with_id_generator(RandomIdGenerator::default())
                .with_resource(Resource::new(vec![
                    KeyValue::new("service.name", service_name.clone()),
                    KeyValue::new("service.version", service_version),
                    KeyValue::new("deployment.environment", environment),
                ])),
        )
        .install_batch(runtime::Tokio)?;

    // Set global tracer provider
    global::set_tracer_provider(tracer.provider().unwrap());

    // Initialize tracing subscriber with OpenTelemetry layer
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::new(
            std::env::var("RUST_LOG").unwrap_or_else(|_| "info".into()),
        ))
        .with(tracing_subscriber::fmt::layer())
        .with(tracing_opentelemetry::layer())
        .init();

    tracing::info!("✅ OpenTelemetry initialized for {}", service_name);
    tracing::info!("📡 Sending traces to: {}", otlp_endpoint);

    Ok(())
}

pub async fn shutdown_telemetry() {
    global::shutdown_tracer_provider();
    tracing::info!("OpenTelemetry shutdown complete");
}
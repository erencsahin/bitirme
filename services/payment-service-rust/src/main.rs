mod config;
mod database;
mod dto;
mod handlers;
mod middleware;
mod models;
mod services;
mod telemetry;

use axum::{
    routing::{get, post},
    Router,
};
use config::Config;
use services::user_client::UserServiceClient;
use std::sync::Arc;
use tower_http::cors::CorsLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Load environment variables
    dotenv::dotenv().ok();

    // Initialize OpenTelemetry tracing
    telemetry::init_telemetry()?;

    // Load configuration
    let config = Arc::new(Config::from_env()?);
    tracing::info!("Configuration loaded successfully");

    // Initialize database
    let db_pool = database::create_pool(&config.database_url).await?;
    tracing::info!("Database connection established");

    // Initialize Redis
    let redis_client = redis::Client::open(config.redis_url.clone())?;
    let redis_conn = redis_client.get_connection_manager().await?;
    tracing::info!("Redis connection established");

    // Initialize User Service client
    let user_service_url = std::env::var("USER_SERVICE_URL")
        .unwrap_or_else(|_| "http://localhost:8083".to_string());
    let user_client = Arc::new(UserServiceClient::new(user_service_url));
    tracing::info!("User Service client initialized");

    // Build application state
    let app_state = Arc::new(services::AppState {
        config: config.clone(),
        db_pool,
        redis_conn,
    });

    // Build router
    let app = Router::new()
        .route("/api/health", get(handlers::health::health_check))
        .route("/api/payments", post(handlers::payment::create_payment))
        .route("/api/payments/:id", get(handlers::payment::get_payment))
        .route("/api/payments/order/:order_id", get(handlers::payment::get_payment_by_order))
        .layer(CorsLayer::permissive())
        .with_state(app_state);

    // Start server
    let addr = format!("0.0.0.0:{}", config.port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    tracing::info!("Server listening on {}", addr);

    axum::serve(listener, app).await?;

    Ok(())
}
use axum::Json;
use serde_json::{json, Value};

#[tracing::instrument(name = "health_check")]
pub async fn health_check() -> Json<Value> {
    tracing::info!("Health check called");
    Json(json!({
        "status": "UP",
        "service": "payment-service",
        "timestamp": chrono::Utc::now().to_rfc3339()
    }))
}
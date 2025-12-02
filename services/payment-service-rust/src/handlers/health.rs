use axum::Json;
use serde_json::{json, Value};

pub async fn health_check() -> Json<Value> {
    Json(json!({
        "status": "UP",
        "service": "payment-service",
        "timestamp": chrono::Utc::now().to_rfc3339()
    }))
}
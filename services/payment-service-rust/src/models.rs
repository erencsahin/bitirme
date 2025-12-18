use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use rust_decimal::Decimal;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct Payment {
    pub id: Uuid,
    pub order_id: Uuid,
    pub user_id: Uuid,
    pub amount: Decimal,
    pub currency: String,
    pub payment_method: String,
    pub payment_status: String,
    pub transaction_id: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PaymentStatus {
    Pending,
    Processing,
    Completed,
    Failed,
    Refunded,
}

impl PaymentStatus {
    pub fn as_str(&self) -> &str {
        match self {
            PaymentStatus::Pending => "PENDING",
            PaymentStatus::Processing => "PROCESSING",
            PaymentStatus::Completed => "COMPLETED",
            PaymentStatus::Failed => "FAILED",
            PaymentStatus::Refunded => "REFUNDED",
        }
    }
}
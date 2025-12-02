use crate::{dto::CreatePaymentRequest, models::{Payment, PaymentStatus}};
use anyhow::Result;
use sqlx::PgPool;
use uuid::Uuid;
use chrono::Utc;

pub async fn create_payment(
    pool: &PgPool,
    request: CreatePaymentRequest,
) -> Result<Payment> {
    // Mock payment processing
    let transaction_id = Uuid::new_v4().to_string();
    let payment_status = PaymentStatus::Completed; // Mock: always success

    let payment = sqlx::query_as::<_, Payment>(
        r#"
        INSERT INTO payments (id, order_id, user_id, amount, currency, payment_method, payment_status, transaction_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING *
        "#,
    )
    .bind(Uuid::new_v4())
    .bind(request.order_id)
    .bind(request.user_id)
    .bind(request.amount)
    .bind(request.currency)
    .bind(request.payment_method)
    .bind(payment_status.as_str())
    .bind(Some(transaction_id))
    .bind(Utc::now())
    .bind(Utc::now())
    .fetch_one(pool)
    .await?;

    Ok(payment)
}

pub async fn get_payment(pool: &PgPool, id: Uuid) -> Result<Payment> {
    let payment = sqlx::query_as::<_, Payment>(
        "SELECT * FROM payments WHERE id = $1"
    )
    .bind(id)
    .fetch_one(pool)
    .await?;

    Ok(payment)
}

pub async fn get_payment_by_order(pool: &PgPool, order_id: Uuid) -> Result<Payment> {
    let payment = sqlx::query_as::<_, Payment>(
        "SELECT * FROM payments WHERE order_id = $1"
    )
    .bind(order_id)
    .fetch_one(pool)
    .await?;

    Ok(payment)
}
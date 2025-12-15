use crate::{
    dto::{ApiResponse, CreatePaymentRequest, PaymentResponse},
    services::{payment_service, AppState},
};
use axum::{
    extract::{Path, State},
    http::StatusCode,
    Json,
};
use std::sync::Arc;
use uuid::Uuid;

#[tracing::instrument(name = "create_payment", skip(state))]
pub async fn create_payment(
    State(state): State<Arc<AppState>>,
    Json(request): Json<CreatePaymentRequest>,
) -> Result<Json<ApiResponse<PaymentResponse>>, StatusCode> {
    tracing::info!("Creating payment for order: {}", request.order_id);
    let payment = payment_service::create_payment(&state.db_pool, request)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let response = PaymentResponse {
        id: payment.id,
        order_id: payment.order_id,
        user_id: payment.user_id,
        amount: payment.amount,
        currency: payment.currency,
        payment_method: payment.payment_method,
        payment_status: payment.payment_status,
        transaction_id: payment.transaction_id,
        created_at: payment.created_at.to_rfc3339(),
        updated_at: payment.updated_at.to_rfc3339(),
    };
    

    Ok(Json(ApiResponse::success(response)))
}

pub async fn get_payment(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<ApiResponse<PaymentResponse>>, StatusCode> {
    let payment = payment_service::get_payment(&state.db_pool, id)
        .await
        .map_err(|_| StatusCode::NOT_FOUND)?;

    let response = PaymentResponse {
        id: payment.id,
        order_id: payment.order_id,
        user_id: payment.user_id,
        amount: payment.amount,
        currency: payment.currency,
        payment_method: payment.payment_method,
        payment_status: payment.payment_status,
        transaction_id: payment.transaction_id,
        created_at: payment.created_at.to_rfc3339(),
        updated_at: payment.updated_at.to_rfc3339(),
    };

    Ok(Json(ApiResponse::success(response)))
}

pub async fn get_payment_by_order(
    State(state): State<Arc<AppState>>,
    Path(order_id): Path<Uuid>,
) -> Result<Json<ApiResponse<PaymentResponse>>, StatusCode> {
    let payment = payment_service::get_payment_by_order(&state.db_pool, order_id)
        .await
        .map_err(|_| StatusCode::NOT_FOUND)?;

    let response = PaymentResponse {
        id: payment.id,
        order_id: payment.order_id,
        user_id: payment.user_id,
        amount: payment.amount,
        currency: payment.currency,
        payment_method: payment.payment_method,
        payment_status: payment.payment_status,
        transaction_id: payment.transaction_id,
        created_at: payment.created_at.to_rfc3339(),
        updated_at: payment.updated_at.to_rfc3339(),
    };

    Ok(Json(ApiResponse::success(response)))
}
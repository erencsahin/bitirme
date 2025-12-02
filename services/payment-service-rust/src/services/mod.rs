use crate::config::Config;
use redis::aio::ConnectionManager;
use sqlx::PgPool;
use std::sync::Arc;

pub mod payment_service;
pub mod user_client;

pub struct AppState {
    pub config: Arc<Config>,
    pub db_pool: PgPool,
    pub redis_conn: ConnectionManager,
}
use crate::handlers;
use crate::services::listing_service::WorkerHandle;
use axum::{Router, routing::post};

pub fn router(worker: WorkerHandle) -> Router {
    Router::new()
        .route("/fetchPrices", post(handlers::listings::fetch_prices))
        .with_state(worker)
}

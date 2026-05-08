use crate::dto::requests::FetchPricesRequest;
use crate::dto::responses::FetchPricesResponse;
use crate::services::listing_service::WorkerHandle;
use axum::{Json, extract::State, http::StatusCode};
use tracing::error;

pub async fn fetch_prices(
    State(worker): State<WorkerHandle>,
    Json(request): Json<FetchPricesRequest>,
) -> Result<Json<FetchPricesResponse>, (StatusCode, String)> {
    match worker.fetch_prices(request).await {
        Ok(response) => Ok(Json(response)),
        Err(err) => {
            error!(?err, "fetchPrices failed");
            Err((StatusCode::INTERNAL_SERVER_ERROR, err.to_string()))
        }
    }
}

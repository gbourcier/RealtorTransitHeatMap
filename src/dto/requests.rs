use serde::Deserialize;

#[derive(Debug, Clone, Deserialize)]
pub struct FetchPricesRequest {
    pub city: String,
    pub min_price: Option<u64>,
    pub max_price: Option<u64>,
}

use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct Listing {
    pub id: String,
    pub price: u64,
    pub address: String,
}

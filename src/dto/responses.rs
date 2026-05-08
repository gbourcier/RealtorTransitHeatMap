use crate::models::listing::Listing;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize)]
pub struct FetchPricesResponse {
    pub listings: Vec<Listing>,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct AsyncPropertySearchResponse {
    pub paging: Paging,
    pub results: Vec<ListingResult>,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct Paging {
    pub records_per_page: u32,
    pub current_page: u32,
    pub total_records: u32,
    pub total_pages: u32,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct ListingResult {
    pub id: String,
    pub mls_number: String,
    pub property: Property,
    #[serde(rename = "RelativeDetailsURL")]
    pub relative_details_url: String,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct Property {
    pub price_unformatted_value: String,
    pub address: Address,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct Address {
    pub address_text: String,
    pub latitude: String,
    pub longitude: String,
}

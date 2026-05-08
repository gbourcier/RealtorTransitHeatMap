use crate::dto::requests::FetchPricesRequest;
use crate::dto::responses::FetchPricesResponse;
use crate::models::listing::Listing;
use tokio::sync::{mpsc, oneshot};
use tracing::{info, warn};

pub struct Job {
    pub request: FetchPricesRequest,
    pub reply: oneshot::Sender<anyhow::Result<FetchPricesResponse>>,
}

#[derive(Clone)]
pub struct WorkerHandle {
    tx: mpsc::Sender<Job>,
}

impl WorkerHandle {
    pub async fn fetch_prices(
        &self,
        request: FetchPricesRequest,
    ) -> anyhow::Result<FetchPricesResponse> {
        let (reply_tx, reply_rx) = oneshot::channel();
        self.tx
            .send(Job { request, reply: reply_tx })
            .await
            .map_err(|_| anyhow::anyhow!("worker has shut down"))?;
        reply_rx
            .await
            .map_err(|_| anyhow::anyhow!("worker dropped reply"))?
    }
}

pub fn spawn() -> WorkerHandle {
    let (tx, mut rx) = mpsc::channel::<Job>(32);
    let http = reqwest::Client::new();

    tokio::spawn(async move {
        info!("worker started");
        while let Some(job) = rx.recv().await {
            let result = handle_job(&http, job.request).await;
            if job.reply.send(result).is_err() {
                warn!("reply receiver dropped before worker responded");
            }
        }
        info!("worker shutting down");
    });

    WorkerHandle { tx }
}

async fn handle_job(
    _http: &reqwest::Client,
    request: FetchPricesRequest,
) -> anyhow::Result<FetchPricesResponse> {
    info!(?request, "handling fetchPrices job");
    // TODO: replace with a real call to realtor.ca and parse the response.
    Ok(FetchPricesResponse {
        listings: vec![Listing {
            id: "stub-1".into(),
            price: 750_000,
            address: format!("123 Fake St, {}", request.city),
        }],
    })
}

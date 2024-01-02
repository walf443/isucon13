use axum::extract::State;
use isupipe_core::services::livestream_comment_service::{
    HaveLivestreamCommentService, LivestreamCommentService,
};
use isupipe_core::services::manager::ServiceManager;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_infra::services::manager::ServiceManagerInfra;

#[derive(Debug, serde::Serialize)]
pub struct PaymentResult {
    total_tip: i64,
}

pub async fn get_payment_result<S: ServiceManager>(
    State(AppState { pool, .. }): State<AppState<S>>,
) -> Result<axum::Json<PaymentResult>, Error> {
    let service = ServiceManagerInfra::new(pool.clone());

    let total_tip = service.livestream_comment_service().get_sum_tip().await?;

    Ok(axum::Json(PaymentResult { total_tip }))
}

use axum::extract::State;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_infra::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;

#[derive(Debug, serde::Serialize)]
pub struct PaymentResult {
    total_tip: i64,
}

pub async fn get_payment_result(
    State(AppState { pool, .. }): State<AppState>,
) -> Result<axum::Json<PaymentResult>, Error> {
    let mut tx = pool.begin().await?;

    let comment_repo = LivestreamCommentRepositoryInfra {};
    let total_tip = comment_repo.get_sum_tip(&mut *tx).await?;

    tx.commit().await?;

    Ok(axum::Json(PaymentResult { total_tip }))
}

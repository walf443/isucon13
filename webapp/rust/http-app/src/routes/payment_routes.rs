use axum::extract::State;
use isupipe_core::models::mysql_decimal::MysqlDecimal;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;

#[derive(Debug, serde::Serialize)]
pub struct PaymentResult {
    total_tip: i64,
}

pub async fn get_payment_result(
    State(AppState { pool, .. }): State<AppState>,
) -> Result<axum::Json<PaymentResult>, Error> {
    let mut tx = pool.begin().await?;

    let MysqlDecimal(total_tip) =
        sqlx::query_scalar("SELECT IFNULL(SUM(tip), 0) FROM livecomments")
            .fetch_one(&mut *tx)
            .await?;

    tx.commit().await?;

    Ok(axum::Json(PaymentResult { total_tip }))
}

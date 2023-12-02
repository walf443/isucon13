use crate::utils::fill_livecomment_report_response;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::livestream::LivestreamModel;
use isupipe_core::models::livestream_comment_report::{LivecommentReport, LivecommentReportModel};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};

pub async fn get_livecomment_reports_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<Vec<LivecommentReport>>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let livestream_model: LivestreamModel =
        sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
            .bind(livestream_id)
            .fetch_one(&mut *tx)
            .await?;

    if livestream_model.user_id != user_id {
        return Err(Error::Forbidden(
            "can't get other streamer's livecomment reports".into(),
        ));
    }

    let report_models: Vec<LivecommentReportModel> =
        sqlx::query_as("SELECT * FROM livecomment_reports WHERE livestream_id = ?")
            .bind(livestream_id)
            .fetch_all(&mut *tx)
            .await?;

    let mut reports = Vec::with_capacity(report_models.len());
    for report_model in report_models {
        let report = fill_livecomment_report_response(&mut tx, report_model).await?;
        reports.push(report);
    }

    tx.commit().await?;

    Ok(axum::Json(reports))
}

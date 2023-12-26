use crate::responses::livestream_comment_report_response::LivestreamCommentReportResponse;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::LivestreamCommentId;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use isupipe_core::services::livestream_comment_report_service::LivestreamCommentReportService;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use isupipe_infra::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_infra::services::LivestreamCommentReportServiceInfra;
use std::sync::Arc;

pub async fn get_livecomment_reports_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<Vec<LivestreamCommentReportResponse>>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let livestream_repo = LivestreamRepositoryInfra {};
    let livestream_model = livestream_repo
        .find(&mut *tx, livestream_id)
        .await?
        .unwrap();

    if livestream_model.user_id.get() != user_id {
        return Err(Error::Forbidden(
            "can't get other streamer's livecomment reports".into(),
        ));
    }

    let report_repo = LivestreamCommentReportRepositoryInfra {};
    let report_models = report_repo
        .find_all_by_livestream_id(&mut *tx, &livestream_model.id)
        .await?;

    let mut reports = Vec::with_capacity(report_models.len());
    for report_model in report_models {
        let report = LivestreamCommentReportResponse::build(&mut tx, report_model).await?;
        reports.push(report);
    }

    tx.commit().await?;

    Ok(axum::Json(reports))
}
pub async fn report_livecomment_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id, livecomment_id)): Path<(i64, i64)>,
) -> Result<(StatusCode, axum::Json<LivestreamCommentReportResponse>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    let livestream_id = LivestreamId::new(livestream_id);
    let comment_report_service = LivestreamCommentReportServiceInfra::new(Arc::new(pool.clone()));
    let comment_id = LivestreamCommentId::new(livecomment_id);
    let report = comment_report_service
        .create(&user_id, &livestream_id, &comment_id)
        .await?;

    let mut conn = pool.acquire().await?;
    let report = LivestreamCommentReportResponse::build(&mut conn, report).await?;

    Ok((StatusCode::CREATED, axum::Json(report)))
}

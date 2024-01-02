use crate::responses::livestream_comment_report_response::LivestreamCommentReportResponse;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::LivestreamCommentId;
use isupipe_core::models::user::UserId;
use isupipe_core::services::livestream_comment_report_service::LivestreamCommentReportService;
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};

pub async fn get_livecomment_reports_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
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
    let livestream_id = LivestreamId::new(livestream_id);

    let livestream_model = service
        .livestream_service()
        .find(&livestream_id)
        .await?
        .unwrap();

    if livestream_model.user_id.get() != user_id {
        return Err(Error::Forbidden(
            "can't get other streamer's livecomment reports".into(),
        ));
    }

    let report_models = service
        .livestream_comment_report_service()
        .find_all_by_livestream_id(&livestream_model.id)
        .await?;

    let mut reports = Vec::with_capacity(report_models.len());
    for report_model in report_models {
        let report =
            LivestreamCommentReportResponse::build_by_service(&service, report_model).await?;
        reports.push(report);
    }

    Ok(axum::Json(reports))
}
pub async fn report_livecomment_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
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
    let comment_id = LivestreamCommentId::new(livecomment_id);
    let report = service
        .livestream_comment_report_service()
        .create(&user_id, &livestream_id, &comment_id)
        .await?;

    let report = LivestreamCommentReportResponse::build_by_service(&service, report).await?;

    Ok((StatusCode::CREATED, axum::Json(report)))
}

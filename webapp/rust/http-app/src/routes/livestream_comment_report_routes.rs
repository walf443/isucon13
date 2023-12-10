use crate::utils::fill_livecomment_report_response;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;
use isupipe_core::models::livestream_comment::LivestreamCommentModel;
use isupipe_core::models::livestream_comment_report::{LivecommentReport, LivecommentReportModel};
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::livestream_repository::LivestreamRepositoryInfra;

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

    let livestream_repo = LivestreamRepositoryInfra {};
    let livestream_model = livestream_repo
        .find(&mut *tx, livestream_id)
        .await?
        .unwrap();

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
pub async fn report_livecomment_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id, livecomment_id)): Path<(i64, i64)>,
) -> Result<(StatusCode, axum::Json<LivecommentReport>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let livestream_repo = LivestreamRepositoryInfra {};
    livestream_repo
        .find(&mut *tx, livestream_id)
        .await?
        .ok_or(Error::NotFound("livestream not found".into()))?;

    let _: LivestreamCommentModel = sqlx::query_as("SELECT * FROM livecomments WHERE id = ?")
        .bind(livecomment_id)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::NotFound("livecomment not found".into()))?;

    let now = Utc::now().timestamp();
    let rs = sqlx::query(
        "INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)",
    )
        .bind(user_id)
        .bind(livestream_id)
        .bind(livecomment_id)
        .bind(now)
        .execute(&mut *tx)
        .await?;
    let report_id = rs.last_insert_id() as i64;

    let report = fill_livecomment_report_response(
        &mut tx,
        LivecommentReportModel {
            id: report_id,
            user_id,
            livestream_id,
            livecomment_id,
            created_at: now,
        },
    )
    .await?;

    tx.commit().await?;

    Ok((StatusCode::CREATED, axum::Json(report)))
}

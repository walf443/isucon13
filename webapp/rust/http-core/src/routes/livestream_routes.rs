use crate::error::Error;
use crate::responses::livestream_response::LivestreamResponse;
use crate::routes::livestream_comment_report_routes::{
    get_livecomment_reports_handler, report_livecomment_handler,
};
use crate::routes::livestream_comment_routes::post_livecomment_handler;
use crate::routes::livestream_reaction_routes::{get_reactions_handler, post_reaction_handler};
use crate::state::AppState;
use crate::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::Router;
use axum_extra::extract::SignedCookieJar;
use chrono::{DateTime, NaiveDate, TimeZone, Utc};
use isupipe_core::models::livestream::{CreateLivestream, Livestream, LivestreamId};
use isupipe_core::models::livestream_statistics::LivestreamStatistics;
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::ng_word::{CreateNgWord, NgWord};
use isupipe_core::models::tag::TagId;
use isupipe_core::models::user::UserId;
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::livestream_statistics_service::LivestreamStatisticsService;
use isupipe_core::services::livestream_viewers_history_service::LivestreamViewersHistoryService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::ng_word_service::NgWordService;
use isupipe_core::services::ServiceError;

// handle /api/livestreams/
pub fn livestreams_routes<S: ServiceManager + 'static>() -> Router<AppState<S>> {
    Router::new()
        .route(
            "/reservation",
            axum::routing::post(reserve_livestream_handler),
        )
        .route("/search", axum::routing::get(search_livestreams_handler))
        .route(
            "/:livestream_id",
            axum::routing::get(get_livestream_handler),
        )
        .nest("/:livestream_id/", livestream_routes())
}

// handle /api/livestream/:livestream_id/ resources
fn livestream_routes<S: ServiceManager + 'static>() -> Router<AppState<S>> {
    Router::new()
        .route(
            "/livecomment",
            axum::routing::get(get_livestream_handler).post(post_livecomment_handler),
        )
        .route(
            "/livecomment/:livecomment_id/report",
            axum::routing::post(report_livecomment_handler),
        )
        .route(
            "/reaction",
            axum::routing::get(get_reactions_handler).post(post_reaction_handler),
        )
        .route(
            "/report",
            axum::routing::get(get_livecomment_reports_handler),
        )
        .route("/ngwords", axum::routing::get(get_ngwords))
        .route("/moderate", axum::routing::post(moderate_handler))
        .route("/enter", axum::routing::post(enter_livestream_handler))
        .route("/exit", axum::routing::post(exit_livestream_handler))
        .route(
            "/statistics",
            axum::routing::get(get_livestream_statistics_handler),
        )
}

#[derive(Debug, serde::Deserialize)]
pub struct ReserveLivestreamRequest {
    tags: Vec<i64>,
    title: String,
    description: String,
    playlist_url: String,
    thumbnail_url: String,
    start_at: i64,
    end_at: i64,
}

pub async fn reserve_livestream_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    axum::Json(req): axum::Json<ReserveLivestreamRequest>,
) -> Result<(StatusCode, axum::Json<LivestreamResponse>), Error> {
    verify_user_session(&jar).await?;

    if req.tags.iter().any(|&tag_id| tag_id > 103) {
        tracing::error!("unexpected tags: {:?}", req);
    }

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    // 2023/11/25 10:00からの１年間の期間内であるかチェック
    let term_start_at = Utc.from_utc_datetime(
        &NaiveDate::from_ymd_opt(2023, 11, 25)
            .unwrap()
            .and_hms_opt(1, 0, 0)
            .unwrap(),
    );
    let term_end_at = Utc.from_utc_datetime(
        &NaiveDate::from_ymd_opt(2024, 11, 25)
            .unwrap()
            .and_hms_opt(1, 0, 0)
            .unwrap(),
    );
    let reserve_start_at = DateTime::from_timestamp(req.start_at, 0).unwrap();
    let reserve_end_at = DateTime::from_timestamp(req.end_at, 0).unwrap();
    if reserve_start_at >= term_end_at || reserve_end_at <= term_start_at {
        return Err(Error::BadRequest("bad reservation time range".into()));
    }

    let mut tag_ids = Vec::with_capacity(req.tags.len());
    for tag_id in req.tags {
        tag_ids.push(TagId::new(tag_id));
    }

    let result = service
        .livestream_service()
        .create(
            &CreateLivestream {
                user_id: user_id.clone(),
                title: req.title.clone(),
                description: req.description.clone(),
                playlist_url: req.playlist_url.clone(),
                thumbnail_url: req.thumbnail_url.clone(),
                start_at: req.start_at,
                end_at: req.end_at,
            },
            &tag_ids,
        )
        .await;

    match result {
        Ok(livestream) => {
            let livestream = LivestreamResponse::build_by_service(&service, &livestream).await?;

            Ok((StatusCode::CREATED, axum::Json(livestream)))
        }
        Err(e) => match e {
            ServiceError::InvalidReservationRange => Err(Error::BadRequest(
                format!(
                    "予約期間 {} ~ {}に対して、予約区間 {} ~ {}が予約できません",
                    term_start_at.timestamp(),
                    term_end_at.timestamp(),
                    req.start_at,
                    req.end_at
                )
                .into(),
            )),
            _ => Err(e.into()),
        },
    }
}

#[derive(Debug, serde::Deserialize)]
pub struct SearchLivestreamsQuery {
    #[serde(default)]
    pub tag: String,
    #[serde(default)]
    pub limit: String,
}

pub async fn search_livestreams_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    Query(SearchLivestreamsQuery {
        tag: key_tag_name,
        limit,
    }): Query<SearchLivestreamsQuery>,
) -> Result<axum::Json<Vec<LivestreamResponse>>, Error> {
    let livestream_models: Vec<Livestream> = if key_tag_name.is_empty() {
        let limit = if limit.is_empty() {
            None
        } else {
            let limit: i64 = limit
                .parse()
                .map_err(|_| Error::BadRequest("failed to parse limit".into()))?;

            Some(limit)
        };

        service
            .livestream_service()
            .find_recent_livestreams(limit)
            .await?
    } else {
        // タグによる取得
        service
            .livestream_service()
            .find_recent_by_tag_name(&key_tag_name)
            .await?
    };

    let livestreams =
        LivestreamResponse::bulk_build_by_service(&service, &livestream_models).await?;

    Ok(axum::Json(livestreams))
}
pub async fn get_my_livestreams_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
) -> Result<axum::Json<Vec<LivestreamResponse>>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    let livestream_models = service
        .livestream_service()
        .find_all_by_user_id(&user_id)
        .await?;

    let livestreams =
        LivestreamResponse::bulk_build_by_service(&service, &livestream_models).await?;

    Ok(axum::Json(livestreams))
}

pub async fn get_livestream_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<LivestreamResponse>, Error> {
    verify_user_session(&jar).await?;
    let livestream_id = LivestreamId::new(livestream_id);

    let livestream_model = service.livestream_service().find(&livestream_id).await?;

    if livestream_model.is_none() {
        return Err(Error::NotFound(
            "not found livestream that has the given id".into(),
        ));
    }
    let livestream_model = livestream_model.unwrap();

    let livestream = LivestreamResponse::build_by_service(&service, &livestream_model).await?;

    Ok(axum::Json(livestream))
}
pub async fn get_ngwords<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<Vec<NgWord>>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);
    let livestream_id = LivestreamId::new(livestream_id);

    let ng_words = service
        .ng_word_service()
        .find_all_by_livestream_id_and_user_id(&livestream_id, &user_id)
        .await?;

    Ok(axum::Json(ng_words))
}
#[derive(Debug, serde::Deserialize)]
pub struct ModerateRequest {
    ng_word: String,
}

#[derive(Debug, serde::Serialize)]
pub struct ModerateResponse {
    word_id: i64,
}

// NGワードを登録
pub async fn moderate_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    axum::Json(req): axum::Json<ModerateRequest>,
) -> Result<(StatusCode, axum::Json<ModerateResponse>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    let livestream_id = LivestreamId::new(livestream_id);

    // 配信者自身の配信に対するmoderateなのかを検証
    let is_exist = service
        .livestream_service()
        .exist_by_id_and_user_id(&livestream_id, &user_id)
        .await?;
    if !is_exist {
        return Err(Error::BadRequest(
            "A streamer can't moderate livestreams that other streamers own".into(),
        ));
    }

    let created_at = Utc::now().timestamp();
    let word_id = service
        .ng_word_service()
        .create(&CreateNgWord {
            user_id: user_id.clone(),
            livestream_id: livestream_id.clone(),
            word: req.ng_word,
            created_at,
        })
        .await?;

    Ok((
        StatusCode::CREATED,
        axum::Json(ModerateResponse {
            word_id: word_id.get(),
        }),
    ))
}

// viewerテーブルの廃止
pub async fn enter_livestream_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<(), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);
    let livestream_id = LivestreamId::new(livestream_id);

    let created_at = Utc::now().timestamp();

    service
        .livestream_viewers_history_service()
        .create(&CreateLivestreamViewersHistory {
            user_id: user_id.clone(),
            livestream_id: livestream_id.clone(),
            created_at,
        })
        .await?;

    Ok(())
}
pub async fn exit_livestream_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<(), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);
    let livestream_id = LivestreamId::new(livestream_id);

    service
        .livestream_viewers_history_service()
        .delete_by_livestream_id_and_user_id(&livestream_id, &user_id)
        .await?;

    Ok(())
}
pub async fn get_livestream_statistics_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<LivestreamStatistics>, Error> {
    verify_user_session(&jar).await?;

    let livestream_id = LivestreamId::new(livestream_id);

    let livestream = service
        .livestream_service()
        .find(&livestream_id)
        .await?
        .ok_or(Error::BadRequest("".into()))?;

    let stats = service
        .livestream_statistics_service()
        .get_stats(&livestream)
        .await?;

    Ok(axum::Json(stats))
}

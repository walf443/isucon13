use crate::responses::livestream_response::LivestreamResponse;
use crate::responses::theme_response::ThemeResponse;
use crate::responses::user_response::UserResponse;
use crate::routes::user_icon_routes::get_icon_handler;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum::routing::get;
use axum::Router;
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::user::UserId;
use isupipe_core::models::user_statistics::UserStatistics;
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::theme_service::ThemeService;
use isupipe_core::services::user_service::UserService;
use isupipe_core::services::user_statistics_service::UserStatisticsService;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};

pub fn user_routes<S: ServiceManager + 'static>() -> Router<AppState<S>> {
    Router::new()
        .route("/me", axum::routing::get(get_me_handler::<S>))
        // フロントエンドで、配信予約のコラボレーターを指定する際に必要
        .route("/:username", axum::routing::get(get_user_handler::<S>))
        .route("/:username/theme", get(get_streamer_theme_handler::<S>))
        .route(
            "/:username/livestream",
            get(get_user_livestreams_handler::<S>),
        )
        .route(
            "/:username/statistics",
            axum::routing::get(get_user_statistics_handler::<S>),
        )
        .route("/:username/icon", axum::routing::get(get_icon_handler::<S>))
}

// 配信者のテーマ取得API
// GET /api/user/:username/theme
pub async fn get_streamer_theme_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<ThemeResponse>, Error> {
    verify_user_session(&jar).await?;

    let user = service
        .user_service()
        .find_by_name(&username)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the given username".into(),
        ))?;

    let theme_model = service.theme_service().find_by_user_id(&user.id).await?;

    Ok(axum::Json(ThemeResponse {
        id: theme_model.id.get(),
        dark_mode: theme_model.dark_mode,
    }))
}
pub async fn get_user_livestreams_handler<S: ServiceManager>(
    State(AppState { service, pool, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<Vec<LivestreamResponse>>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user = service
        .user_service()
        .find_by_name(&username)
        .await?
        .ok_or(Error::NotFound("user not found".into()))?;

    let livestream_models = service
        .livestream_service()
        .find_all_by_user_id(&user.id)
        .await?;

    let mut livestreams = Vec::with_capacity(livestream_models.len());
    for livestream_model in livestream_models {
        let livestream = LivestreamResponse::build(&mut tx, livestream_model).await?;
        livestreams.push(livestream);
    }

    tx.commit().await?;

    Ok(axum::Json(livestreams))
}
pub async fn get_me_handler<S: ServiceManager>(
    State(AppState { service, pool, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
) -> Result<axum::Json<UserResponse>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    let user_model = service
        .user_service()
        .find(&user_id)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the userid in session".into(),
        ))?;

    let mut tx = pool.begin().await?;

    let user = UserResponse::build(&mut tx, user_model).await?;

    tx.commit().await?;

    Ok(axum::Json(user))
}

// ユーザ詳細API
// GET /api/user/:username
pub async fn get_user_handler<S: ServiceManager>(
    State(AppState { service, pool, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<UserResponse>, Error> {
    verify_user_session(&jar).await?;

    let user_model = service
        .user_service()
        .find_by_name(&username)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the given username".into(),
        ))?;

    let mut tx = pool.begin().await?;

    let user = UserResponse::build(&mut tx, user_model).await?;

    tx.commit().await?;

    Ok(axum::Json(user))
}

pub async fn get_user_statistics_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<UserStatistics>, Error> {
    verify_user_session(&jar).await?;

    let user = service
        .user_service()
        .find_by_name(&username)
        .await?
        .ok_or(Error::BadRequest("".into()))?;

    let stats = service.user_statistics_service().get_stats(&user).await?;

    Ok(axum::Json(stats))
}

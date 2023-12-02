use crate::utils::{fill_livestream_response, fill_user_response};
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::livestream::{Livestream, LivestreamModel};
use isupipe_core::models::theme::{Theme, ThemeModel};
use isupipe_core::models::user::{User, UserModel};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};

// 配信者のテーマ取得API
// GET /api/user/:username/theme
pub async fn get_streamer_theme_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<Theme>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user_id: i64 = sqlx::query_scalar("SELECT id FROM users WHERE name = ?")
        .bind(username)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the given username".into(),
        ))?;

    let theme_model: ThemeModel = sqlx::query_as("SELECT * FROM themes WHERE user_id = ?")
        .bind(user_id)
        .fetch_one(&mut *tx)
        .await?;

    tx.commit().await?;

    Ok(axum::Json(Theme {
        id: theme_model.id,
        dark_mode: theme_model.dark_mode,
    }))
}
pub async fn get_user_livestreams_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<Vec<Livestream>>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user: UserModel = sqlx::query_as("SELECT * FROM users WHERE name = ?")
        .bind(username)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::NotFound("user not found".into()))?;

    let livestream_models: Vec<LivestreamModel> =
        sqlx::query_as("SELECT * FROM livestreams WHERE user_id = ?")
            .bind(user.id)
            .fetch_all(&mut *tx)
            .await?;
    let mut livestreams = Vec::with_capacity(livestream_models.len());
    for livestream_model in livestream_models {
        let livestream = fill_livestream_response(&mut tx, livestream_model).await?;
        livestreams.push(livestream);
    }

    tx.commit().await?;

    Ok(axum::Json(livestreams))
}
pub async fn get_me_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
) -> Result<axum::Json<User>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let user_model: UserModel = sqlx::query_as("SELECT * FROM users WHERE id = ?")
        .bind(user_id)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the userid in session".into(),
        ))?;

    let user = fill_user_response(&mut tx, user_model).await?;

    tx.commit().await?;

    Ok(axum::Json(user))
}

// ユーザ詳細API
// GET /api/user/:username
pub async fn get_user_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<User>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user_model: UserModel = sqlx::query_as("SELECT * FROM users WHERE name = ?")
        .bind(username)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the given username".into(),
        ))?;

    let user = fill_user_response(&mut tx, user_model).await?;

    tx.commit().await?;

    Ok(axum::Json(user))
}

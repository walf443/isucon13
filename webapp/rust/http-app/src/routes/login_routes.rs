use async_session::{CookieStore, SessionStore};
use axum::extract::State;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;
use uuid::Uuid;
use isupipe_core::models::user::UserModel;
use isupipe_http_core::{DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY, DEFAULT_USERNAME_KEY, DEFUALT_SESSION_EXPIRES_KEY};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct LoginRequest {
    username: String,
    // password is non-hashed password.
    password: String,
}
// ユーザログインAPI
// POST /api/login
pub async fn login_handler(
    State(AppState { pool, .. }): State<AppState>,
    mut jar: SignedCookieJar,
    axum::Json(req): axum::Json<LoginRequest>,
) -> Result<(SignedCookieJar, ()), Error> {
    let mut tx = pool.begin().await?;

    // usernameはUNIQUEなので、whereで一意に特定できる
    let user_model: UserModel = sqlx::query_as("SELECT * FROM users WHERE name = ?")
        .bind(req.username)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::Unauthorized("invalid username or password".into()))?;

    tx.commit().await?;

    let hashed_password = user_model.hashed_password.unwrap();
    if !bcrypt::verify(&req.password, &hashed_password)? {
        return Err(Error::Unauthorized("invalid username or password".into()));
    }

    let session_end_at = Utc::now() + chrono::Duration::hours(1);
    let session_id = Uuid::new_v4().to_string();
    let mut sess = async_session::Session::new();
    sess.insert(DEFAULT_SESSION_ID_KEY, session_id).unwrap();
    sess.insert(DEFAULT_USER_ID_KEY, user_model.id).unwrap();
    sess.insert(DEFAULT_USERNAME_KEY, user_model.name).unwrap();
    sess.insert(DEFUALT_SESSION_EXPIRES_KEY, session_end_at.timestamp())
        .unwrap();
    let cookie_store = CookieStore::new();
    if let Some(cookie_value) = cookie_store.store_session(sess).await? {
        let cookie =
            axum_extra::extract::cookie::Cookie::build(DEFAULT_SESSION_ID_KEY, cookie_value)
                .domain("u.isucon.dev")
                .max_age(time::Duration::minutes(1000))
                .path("/")
                .finish();
        jar = jar.add(cookie);
    }

    Ok((jar, ()))
}

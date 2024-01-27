use crate::error::Error;
use crate::state::AppState;
use crate::{
    DEFAULT_SESSION_ID_KEY, DEFAULT_USERNAME_KEY, DEFAULT_USER_ID_KEY, DEFUALT_SESSION_EXPIRES_KEY,
};
use async_session::{CookieStore, SessionStore};
use axum::extract::State;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::user_service::UserService;
use uuid::Uuid;

#[derive(Debug, serde::Deserialize)]
pub struct LoginRequest {
    username: String,
    // password is non-hashed password.
    password: String,
}
// ユーザログインAPI
// POST /api/login
pub async fn login_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    mut jar: SignedCookieJar,
    axum::Json(req): axum::Json<LoginRequest>,
) -> Result<(SignedCookieJar, ()), Error> {
    let user_model = service
        .user_service()
        .find_by_name(&req.username)
        .await?
        .ok_or(Error::Unauthorized("invalid username or password".into()))?;

    let hashed_password = user_model.hashed_password.unwrap();
    if !bcrypt::verify(&req.password, &hashed_password)? {
        return Err(Error::Unauthorized("invalid username or password".into()));
    }

    let session_end_at = Utc::now() + chrono::Duration::hours(1);
    let session_id = Uuid::new_v4().to_string();
    let mut sess = async_session::Session::new();
    sess.insert(DEFAULT_SESSION_ID_KEY, session_id).unwrap();
    sess.insert(DEFAULT_USER_ID_KEY, user_model.id).unwrap();
    sess.insert(DEFAULT_USERNAME_KEY, user_model.name.inner())
        .unwrap();
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

use crate::DEFAULT_SESSION_ID_KEY;
use crate::error::Error;
use crate::session::UserSession;
use crate::state::AppState;
use axum::extract::State;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::user_service::UserService;

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
    let sess = UserSession::new(
        user_model.id,
        user_model.name.inner().clone(),
        session_end_at,
    );
    let cookie = axum_extra::extract::cookie::Cookie::build((
        DEFAULT_SESSION_ID_KEY,
        sess.to_cookie_value()?,
    ))
    .domain("u.isucon.dev")
    .max_age(time::Duration::minutes(1000))
    .path("/");
    jar = jar.add(cookie);

    Ok((jar, ()))
}

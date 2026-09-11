use crate::error::Error;
use crate::session::UserSession;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;

pub mod error;
pub mod responses;
pub mod routes;
pub mod session;
pub mod state;

pub const DEFAULT_SESSION_ID_KEY: &str = "SESSIONID";
pub const FALLBACK_IMAGE: &str = "../img/NoImage.jpg";

/// cookie のセッションを検証して返す。cookie が無い・壊れている場合は 403、期限切れは 401。
pub fn verify_user_session(jar: &SignedCookieJar) -> Result<UserSession, Error> {
    let sess = UserSession::load(jar).ok_or(Error::Forbidden("".into()))?;
    if sess.is_expired_at(Utc::now()) {
        return Err(Error::Unauthorized("session has expired".into()));
    }
    Ok(sess)
}

use crate::error::Error;
use async_session::{CookieStore, SessionStore};
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;

pub mod error;
pub mod state;

pub const DEFAULT_SESSION_ID_KEY: &str = "SESSIONID";
pub const DEFUALT_SESSION_EXPIRES_KEY: &str = "EXPIRES";

pub const DEFAULT_USER_ID_KEY: &str = "USERID";
pub const DEFAULT_USERNAME_KEY: &str = "USERNAME";
pub const FALLBACK_IMAGE: &str = "../img/NoImage.jpg";

pub async fn verify_user_session(jar: &SignedCookieJar) -> Result<(), Error> {
    let cookie = jar
        .get(DEFAULT_SESSION_ID_KEY)
        .ok_or(Error::Forbidden("".into()))?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::Forbidden("".into()))?;
    let session_expires: i64 = sess
        .get(DEFUALT_SESSION_EXPIRES_KEY)
        .ok_or(Error::Forbidden("".into()))?;
    let now = Utc::now();
    if now.timestamp() > session_expires {
        return Err(Error::Unauthorized("session has expired".into()));
    }
    Ok(())
}

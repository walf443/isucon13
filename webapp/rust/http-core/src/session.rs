//! 署名付き cookie に載せるログインセッション。
//!
//! サーバー側には何も保存せず、`SignedCookieJar` で署名した cookie の値として
//! `base64(JSON)` を持ち回る。改ざん検知は `SignedCookieJar` 側が担う。
use crate::DEFAULT_SESSION_ID_KEY;
use crate::error::Error;
use axum_extra::extract::SignedCookieJar;
use base64::Engine as _;
use base64::engine::general_purpose::URL_SAFE_NO_PAD;
use chrono::{DateTime, Utc};
use isupipe_core::models::user::UserId;

#[derive(Debug, Clone, PartialEq, serde::Serialize, serde::Deserialize)]
pub struct UserSession {
    pub session_id: String,
    pub user_id: UserId,
    pub username: String,
    /// 有効期限 (unix time, 秒)
    pub expires: i64,
}

impl UserSession {
    pub fn new(user_id: UserId, username: String, expires_at: DateTime<Utc>) -> Self {
        Self {
            session_id: uuid::Uuid::new_v4().to_string(),
            user_id,
            username,
            expires: expires_at.timestamp(),
        }
    }

    pub fn is_expired_at(&self, now: DateTime<Utc>) -> bool {
        now.timestamp() > self.expires
    }

    pub fn to_cookie_value(&self) -> Result<String, Error> {
        let json = serde_json::to_vec(self)?;
        Ok(URL_SAFE_NO_PAD.encode(json))
    }

    pub fn from_cookie_value(value: &str) -> Option<Self> {
        let json = URL_SAFE_NO_PAD.decode(value).ok()?;
        serde_json::from_slice(&json).ok()
    }

    /// cookie からセッションを読み出す。cookie が無い・壊れている場合は `None`。
    pub fn load(jar: &SignedCookieJar) -> Option<Self> {
        let cookie = jar.get(DEFAULT_SESSION_ID_KEY)?;
        Self::from_cookie_value(cookie.value())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn cookie_value_round_trip() {
        let sess = UserSession::new(UserId::new(42), "alice".to_string(), Utc::now());
        let value = sess.to_cookie_value().unwrap();
        assert_eq!(UserSession::from_cookie_value(&value), Some(sess));
    }

    #[test]
    fn broken_cookie_value_is_none() {
        assert_eq!(UserSession::from_cookie_value("not base64!"), None);
        assert_eq!(
            UserSession::from_cookie_value(&URL_SAFE_NO_PAD.encode(b"{}")),
            None
        );
    }

    #[test]
    fn expiry() {
        let now = Utc::now();
        let sess = UserSession::new(UserId::new(1), "a".to_string(), now);
        assert!(!sess.is_expired_at(now));
        assert!(sess.is_expired_at(now + chrono::Duration::seconds(1)));
    }
}

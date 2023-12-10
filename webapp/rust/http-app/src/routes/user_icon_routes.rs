use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use isupipe_core::repos::icon_repository::IconRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{
    verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY, FALLBACK_IMAGE,
};
use isupipe_infra::repos::icon_repository::IconRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;

pub async fn get_icon_handler(
    State(AppState { pool, .. }): State<AppState>,
    Path((username,)): Path<(String,)>,
) -> Result<axum::response::Response, Error> {
    use axum::response::IntoResponse as _;

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user = user_repo.find_by_name(&mut *tx, &username).await?.unwrap();

    let icon_repo = IconRepositoryInfra {};
    let image = icon_repo.find_image_by_user_id(&mut *tx, user.id).await?;

    let headers = [(axum::http::header::CONTENT_TYPE, "image/jpeg")];
    if let Some(image) = image {
        Ok((headers, image).into_response())
    } else {
        let file = tokio::fs::File::open(FALLBACK_IMAGE).await.unwrap();
        let stream = tokio_util::io::ReaderStream::new(file);
        let body = axum::body::StreamBody::new(stream);

        Ok((headers, body).into_response())
    }
}

#[derive(Debug, serde::Deserialize)]
pub struct PostIconRequest {
    #[serde(deserialize_with = "from_base64")]
    image: Vec<u8>,
}
fn from_base64<'de, D>(deserializer: D) -> Result<Vec<u8>, D::Error>
where
    D: serde::Deserializer<'de>,
{
    use base64::Engine as _;
    use serde::de::{Deserialize as _, Error as _};
    let value = String::deserialize(deserializer)?;
    base64::engine::general_purpose::STANDARD
        .decode(value)
        .map_err(D::Error::custom)
}

#[derive(Debug, serde::Serialize)]
pub struct PostIconResponse {
    id: i64,
}

pub async fn post_icon_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    axum::Json(req): axum::Json<PostIconRequest>,
) -> Result<(StatusCode, axum::Json<PostIconResponse>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    sqlx::query("DELETE FROM icons WHERE user_id = ?")
        .bind(user_id)
        .execute(&mut *tx)
        .await?;

    let rs = sqlx::query("INSERT INTO icons (user_id, image) VALUES (?, ?)")
        .bind(user_id)
        .bind(req.image)
        .execute(&mut *tx)
        .await?;
    let icon_id = rs.last_insert_id() as i64;

    tx.commit().await?;

    Ok((
        StatusCode::CREATED,
        axum::Json(PostIconResponse { id: icon_id }),
    ))
}

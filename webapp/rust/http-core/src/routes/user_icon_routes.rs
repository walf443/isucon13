use crate::error::Error;
use crate::state::AppState;
use crate::{FALLBACK_IMAGE, verify_user_session};
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use isupipe_core::services::icon_service::IconService;
use isupipe_core::services::manager::ServiceManager;

pub async fn get_icon_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    Path((username,)): Path<(String,)>,
) -> Result<axum::response::Response, Error> {
    use axum::response::IntoResponse as _;

    let image = service
        .icon_service()
        .find_image_by_user_name(&username)
        .await?;

    let headers = [(axum::http::header::CONTENT_TYPE, "image/jpeg")];
    if let Some(image) = image {
        Ok((headers, image).into_response())
    } else {
        let file = tokio::fs::File::open(FALLBACK_IMAGE).await.unwrap();
        let stream = tokio_util::io::ReaderStream::new(file);
        let body = axum::body::Body::from_stream(stream);

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

pub async fn post_icon_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    axum::Json(req): axum::Json<PostIconRequest>,
) -> Result<(StatusCode, axum::Json<PostIconResponse>), Error> {
    let sess = verify_user_session(&jar)?;

    let user_id = sess.user_id;

    let icon_id = service
        .icon_service()
        .replace_new_image(&user_id, &req.image)
        .await?;

    Ok((
        StatusCode::CREATED,
        axum::Json(PostIconResponse { id: icon_id }),
    ))
}

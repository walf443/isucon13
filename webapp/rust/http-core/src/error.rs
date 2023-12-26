use crate::responses::ResponseError;
use axum::http::StatusCode;
use isupipe_core::repos::ReposError;
use isupipe_core::services::ServiceError;
use isupipe_core::utils::UtilError;
use std::borrow::Cow;

#[derive(Debug, thiserror::Error)]
pub enum Error {
    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),
    #[error("SQLx error: {0}")]
    Sqlx(#[from] sqlx::Error),
    #[error("Repos error: {0}")]
    ReposError(#[from] ReposError),
    #[error("response error: {0}")]
    ResponseError(#[from] ResponseError),
    #[error("utils error: {0}")]
    UtilsError(#[from] UtilError),
    #[error("bcrypt error: {0}")]
    Bcrypt(#[from] bcrypt::BcryptError),
    #[error("async-session error: {0}")]
    AsyncSession(#[from] async_session::Error),
    #[error("{0}")]
    BadRequest(Cow<'static, str>),
    #[error("session error")]
    SessionError,
    #[error("unauthorized: {0}")]
    Unauthorized(Cow<'static, str>),
    #[error("forbidden: {0}")]
    Forbidden(Cow<'static, str>),
    #[error("not found: {0}")]
    NotFound(Cow<'static, str>),
    #[error("{0}")]
    InternalServerError(String),
}

impl From<ServiceError> for Error {
    fn from(err: ServiceError) -> Self {
        match err {
            ServiceError::NotFoundLivestream => Self::NotFound(Cow::from("livestream not found")),
            ServiceError::NotFoundLivestreamComment => {
                Self::NotFound(Cow::from("livecomment not found"))
            }
            e => e.into(),
        }
    }
}

impl axum::response::IntoResponse for Error {
    fn into_response(self) -> axum::response::Response {
        #[derive(Debug, serde::Serialize)]
        struct ErrorResponse {
            error: String,
        }

        let status = match self {
            Self::BadRequest(_) => StatusCode::BAD_REQUEST,
            Self::Unauthorized(_) | Self::SessionError => StatusCode::UNAUTHORIZED,
            Self::Forbidden(_) => StatusCode::FORBIDDEN,
            Self::NotFound(_) => StatusCode::NOT_FOUND,
            Self::Io(_)
            | Self::Sqlx(_)
            | Self::ReposError(_)
            | Self::ResponseError(_)
            | Self::UtilsError(_)
            | Self::Bcrypt(_)
            | Self::AsyncSession(_)
            | Self::InternalServerError(_) => StatusCode::INTERNAL_SERVER_ERROR,
        };

        tracing::error!("{}", self);
        (
            status,
            axum::Json(ErrorResponse {
                error: format!("{}", self),
            }),
        )
            .into_response()
    }
}

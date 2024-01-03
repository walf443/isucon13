pub mod livestream_comment_report_response;
pub mod livestream_comment_response;
pub mod livestream_response;
pub mod reaction_response;
pub mod tag_response;
pub mod theme_response;
pub mod user_response;

use isupipe_core::repos;
use isupipe_core::services::ServiceError;

#[derive(Debug, thiserror::Error)]
pub enum ResponseError {
    #[error("SQLx error: {0}")]
    Sqlx(#[from] sqlx::Error),
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
    #[error("Repos error: {0}")]
    Repos(#[from] repos::ReposError),
    #[error("Service error: {0}")]
    Service(#[from] ServiceError),
}

pub type ResponseResult<T> = Result<T, ResponseError>;

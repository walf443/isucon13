use crate::repos::ReposError;
use thiserror::Error;

pub mod icon_service;
pub mod livestream_comment_report_service;
pub mod livestream_comment_service;
pub mod livestream_service;
pub mod manager;
pub mod reaction_service;
pub mod tag_service;

#[derive(Error, Debug)]
pub enum ServiceError {
    #[error("livestream not found")]
    NotFoundLivestream,
    #[error("livecomment not found")]
    NotFoundLivestreamComment,
    #[error("repos error: #{0}")]
    ReposError(#[from] ReposError),
    #[error("sqlx error: #{0}")]
    SqlxError(#[from] sqlx::Error),
}

pub type ServiceResult<T> = Result<T, ServiceError>;

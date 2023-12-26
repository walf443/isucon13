use crate::repos::ReposError;
use thiserror::Error;

pub mod livestream_comment_report_service;
pub mod manager;

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

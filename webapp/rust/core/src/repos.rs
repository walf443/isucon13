use thiserror::Error;

pub mod livestream_comment_report;
pub mod livestream_repository;
pub mod livestream_viewers_history_repository;
pub mod tag_repository;

#[derive(Debug, Error)]
pub enum ReposError {
    #[error("SQLx error: {0}")]
    Sqlx(#[from] sqlx::Error),
}

pub type Result<T> = std::result::Result<T, ReposError>;

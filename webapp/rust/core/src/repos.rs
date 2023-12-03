use thiserror::Error;

pub mod tag_repository;

#[derive(Debug, Error)]
pub enum ReposError {
    #[error("SQLx error: {0}")]
    Sqlx(#[from] sqlx::Error),
}

pub type Result<T> = std::result::Result<T, ReposError>;

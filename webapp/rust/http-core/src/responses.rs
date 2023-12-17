use isupipe_core::repos;

#[derive(Debug, thiserror::Error)]
pub enum ResponseError {
    #[error("SQLx error: {0}")]
    Sqlx(#[from] sqlx::Error),
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
    #[error("Repos error: {0}")]
    Repos(#[from] repos::ReposError),
}

pub type ResponseResult<T> = Result<T, ResponseError>;

// TODO: あとでどこに配置するべきか考える

use crate::repos;

#[derive(Debug, thiserror::Error)]
pub enum UtilError {
    #[error("SQLx error: {0}")]
    Sqlx(#[from] sqlx::Error),
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
    #[error("Repos error: {0}")]
    Repos(#[from] repos::ReposError),
}

pub type UtilResult<T> = Result<T, UtilError>;

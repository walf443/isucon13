use crate::commands::CommandError;
use crate::repos::ReposError;
use thiserror::Error;

pub mod icon_service;
pub mod initialize_service;
pub mod livestream_comment_report_service;
pub mod livestream_comment_service;
pub mod livestream_service;
pub mod livestream_statistics_service;
pub mod livestream_tag_service;
pub mod livestream_viewers_history_service;
pub mod manager;
pub mod ng_word_service;
pub mod reaction_service;
pub mod tag_service;
pub mod theme_service;
pub mod user_service;
pub mod user_statistics_service;

#[derive(Error, Debug)]
pub enum ServiceError {
    #[error("livestream not found")]
    NotFoundLivestream,
    #[error("livecomment not found")]
    NotFoundLivestreamComment,
    #[error("this comment matched spam")]
    CommentMatchSpam,
    #[error("repos error: #{0}")]
    ReposError(#[from] ReposError),
    #[error("sqlx error: #{0}")]
    SqlxError(#[from] sqlx::Error),
    #[error("command error: #{0}")]
    CommandError(#[from] CommandError),
}

pub type ServiceResult<T> = Result<T, ServiceError>;

//! リポジトリの trait 定義。
//!
//! 各メソッドはクエリ実行ハンドルを `conn: &'c mut DBConn<'c>` の形で受け取る。
//! `DBConn<'c>` は `dyn toasty::Executor + 'c` で、`Db` / `Connection` / `Transaction`
//! のいずれも渡せる。ライフタイム `'c` を明示しているのは、`mockall::automock` と
//! `async_trait` の組み合わせでは省略形 (`&mut dyn Executor`) や `'_` が
//! 受け付けられないため。詳細は [`crate::db::DBConn`] を参照。

use bcrypt::BcryptError;
use thiserror::Error;

pub mod icon_repository;
pub mod livestream_comment_report_repository;
pub mod livestream_comment_repository;
pub mod livestream_repository;
pub mod livestream_tag_repository;
pub mod livestream_viewers_history_repository;
pub mod manager;
pub mod ng_word_repository;
pub mod reaction_repository;
pub mod reservation_slot_repository;
pub mod tag_repository;
pub mod theme_repository;
pub mod user_repository;

#[derive(Debug, Error)]
pub enum ReposError {
    #[error("toasty error: {0}")]
    Toasty(#[from] toasty::Error),
    #[error("bcrypt error: {0}")]
    Bcrypt(#[from] BcryptError),
    #[error("test error")]
    TestError,
}

pub type Result<T> = std::result::Result<T, ReposError>;

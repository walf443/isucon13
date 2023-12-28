use crate::db::DBConn;
use crate::models::theme::Theme;
use crate::models::user::UserId;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait ThemeRepository {
    async fn insert(&self, conn: &mut DBConn, user_id: &UserId, dark_mode: bool) -> Result<()>;
    async fn find_by_user_id(&self, conn: &mut DBConn, user_id: &UserId) -> Result<Theme>;
}

pub trait HaveThemeRepository {
    type Repo: ThemeRepository;

    fn theme_repo(&self) -> &Self::Repo;
}

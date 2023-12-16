use crate::db::DBConn;
use crate::models::theme::ThemeModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait ThemeRepository {
    async fn insert(&self, conn: &mut DBConn, user_id: i64, dark_mode: bool) -> Result<()>;
    async fn find_by_user_id(&self, conn: &mut DBConn, user_id: i64) -> Result<ThemeModel>;
}

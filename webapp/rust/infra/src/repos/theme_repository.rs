#[cfg(test)]
mod create;
#[cfg(test)]
mod find_by_user_id;

use crate::tables::theme::ThemeRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::theme::Theme;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::theme_repository::ThemeRepository;

#[derive(Clone)]
pub struct ThemeRepositoryInfra {}

#[async_trait]
impl ThemeRepository for ThemeRepositoryInfra {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user_id: &UserId,
        dark_mode: bool,
    ) -> isupipe_core::repos::Result<()> {
        ThemeRow::create()
            .user_id(user_id.inner())
            .dark_mode(dark_mode)
            .exec(conn)
            .await?;

        Ok(())
    }

    async fn find_by_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Theme> {
        let row = ThemeRow::filter(ThemeRow::fields().user_id().eq(user_id.inner()))
            .one()
            .exec(conn)
            .await?;

        Ok(row.into())
    }
}

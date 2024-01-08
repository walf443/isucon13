#[cfg(test)]
mod create;
#[cfg(test)]
mod find_by_user_id;

use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::theme::Theme;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::theme_repository::ThemeRepository;

#[derive(Clone)]
pub struct ThemeRepositoryInfra {}

#[async_trait]
impl ThemeRepository for ThemeRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
        dark_mode: bool,
    ) -> isupipe_core::repos::Result<()> {
        sqlx::query!(
            "INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)",
            user_id,
            dark_mode
        )
        .execute(conn)
        .await?;

        Ok(())
    }

    async fn find_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Theme> {
        let theme_model: Theme = sqlx::query_as!(
            Theme,
            "SELECT id, user_id, dark_mode as `dark_mode:bool` FROM themes WHERE user_id = ?",
            user_id
        )
        .fetch_one(conn)
        .await?;

        Ok(theme_model)
    }
}

use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::theme::Theme;
use isupipe_core::repos::theme_repository::ThemeRepository;

pub struct ThemeRepositoryInfra {}

#[async_trait]
impl ThemeRepository for ThemeRepositoryInfra {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        dark_mode: bool,
    ) -> isupipe_core::repos::Result<()> {
        sqlx::query("INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)")
            .bind(user_id)
            .bind(dark_mode)
            .execute(conn)
            .await?;

        Ok(())
    }

    async fn find_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> isupipe_core::repos::Result<Theme> {
        let theme_model: Theme = sqlx::query_as("SELECT * FROM themes WHERE user_id = ?")
            .bind(user_id)
            .fetch_one(conn)
            .await?;

        Ok(theme_model)
    }
}

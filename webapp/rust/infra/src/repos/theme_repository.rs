use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::theme::ThemeModel;
use isupipe_core::repos::theme_repository::ThemeRepository;

pub struct ThemeRepositoryInfra {}

#[async_trait]
impl ThemeRepository for ThemeRepositoryInfra {
    async fn find_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> isupipe_core::repos::Result<ThemeModel> {
        let theme_model: ThemeModel = sqlx::query_as("SELECT * FROM themes WHERE user_id = ?")
            .bind(user_id)
            .fetch_one(conn)
            .await?;

        Ok(theme_model)
    }
}

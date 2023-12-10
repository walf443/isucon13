use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::repos::icon_repository::IconRepository;

pub struct IconRepositoryInfra {}

#[async_trait]
impl IconRepository for IconRepositoryInfra {
    async fn find_image_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> isupipe_core::repos::Result<Option<Vec<u8>>> {
        let image: Option<Vec<u8>> =
            sqlx::query_scalar("SELECT image FROM icons WHERE user_id = ?")
                .bind(user_id)
                .fetch_optional(conn)
                .await?;

        Ok(image)
    }
}

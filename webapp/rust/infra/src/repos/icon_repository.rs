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

    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        image: &Vec<u8>,
    ) -> isupipe_core::repos::Result<i64> {
        let rs = sqlx::query("INSERT INTO icons (user_id, image) VALUES (?, ?)")
            .bind(user_id)
            .bind(image)
            .execute(conn)
            .await?;
        let icon_id = rs.last_insert_id() as i64;

        Ok(icon_id)
    }

    async fn delete_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> isupipe_core::repos::Result<()> {
        sqlx::query("DELETE FROM icons WHERE user_id = ?")
            .bind(user_id)
            .execute(conn)
            .await?;

        Ok(())
    }
}

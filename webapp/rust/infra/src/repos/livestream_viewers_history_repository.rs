use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
use sqlx::Acquire;
pub struct LivestreamViewersHistoryRepositoryInfra {}

#[async_trait]
impl LivestreamViewersHistoryRepository for LivestreamViewersHistoryRepositoryInfra {
    async fn delete_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        user_id: i64,
    ) -> isupipe_core::repos::Result<()> {
        let mut tx = conn.begin().await?;

        sqlx::query(
            "DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?",
        )
        .bind(user_id)
        .bind(livestream_id)
        .execute(&mut *tx)
        .await?;

        tx.commit().await?;

        Ok(())
    }
}

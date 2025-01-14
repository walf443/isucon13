use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
use sqlx::Acquire;

#[derive(Clone)]
pub struct LivestreamViewersHistoryRepositoryInfra {}

#[async_trait]
impl LivestreamViewersHistoryRepository for LivestreamViewersHistoryRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        history: &CreateLivestreamViewersHistory,
    ) -> isupipe_core::repos::Result<()> {
        let mut tx = conn.begin().await?;

        sqlx::query(
            "INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(?, ?, ?)",
        )
            .bind(&history.user_id)
            .bind(&history.livestream_id)
            .bind(history.created_at)
            .execute(&mut *tx)
            .await?;

        tx.commit().await?;

        Ok(())
    }

    async fn count_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let viewers_count = sqlx::query_scalar("SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?")
            .bind(livestream_id)
            .fetch_one(&mut *conn)
            .await?;

        Ok(viewers_count)
    }

    async fn delete_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
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

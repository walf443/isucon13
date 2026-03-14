#[cfg(test)]
mod count_by_livestream_id;
#[cfg(test)]
mod delete_by_livestream_id_and_user_id;

use crate::sqipe_support::bind_sqipe_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_viewers_history::TABLE_LIVESTREAM_VIEWERS_HISTORY;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
use sqipe::{aggregate, table};
use sqipe_mysql::sqipe;
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
        let livestream = &TABLE_LIVESTREAMS;
        let viewers_history = &TABLE_LIVESTREAM_VIEWERS_HISTORY;
        let mut q = sqipe(livestream.table_name());
        q.as_("l");
        q.join(
            viewers_history.table_name(),
            viewers_history.table().col("livestream_id").eq_col(table("l").col("id")),
        );
        q.and_where(table("l").col("id").eq(*livestream_id.inner()));
        q.aggregate(&[aggregate::count_all()]);
        let (sql, binds) = q.to_sql();
        let viewers_count = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
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

        let viewers_history = &TABLE_LIVESTREAM_VIEWERS_HISTORY;
        let mut d = sqipe(viewers_history.table_name()).delete();
        d.and_where(viewers_history.user_id().eq(*user_id.inner()));
        d.and_where(viewers_history.livestream_id().eq(*livestream_id.inner()));
        let (sql, binds) = d.to_sql();
        bind_sqipe_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await?;

        tx.commit().await?;

        Ok(())
    }
}

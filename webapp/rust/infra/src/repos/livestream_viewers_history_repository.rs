#[cfg(test)]
mod count_by_livestream_id;
#[cfg(test)]
mod delete_by_livestream_id_and_user_id;

use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_viewers_history::TABLE_LIVESTREAM_VIEWERS_HISTORY;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
use qbey::RawSql;
use qbey_mysql::qbey;
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
        let mut q = qbey(livestream.table());
        q.as_("l");
        let l = livestream.as_("l");
        q.join(
            viewers_history.table(),
            viewers_history.livestream_id().eq_col(l.id()),
        );
        q.and_where(l.id().eq(*livestream_id.inner()));
        q.add_select_expr(RawSql::new("COUNT(*)"), None);
        let (sql, binds) = q.to_sql();
        let viewers_count = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
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
        let mut d = qbey(viewers_history.table()).into_delete();
        d.and_where(viewers_history.user_id().eq(*user_id.inner()));
        d.and_where(viewers_history.livestream_id().eq(*livestream_id.inner()));
        let (sql, binds) = d.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await?;

        tx.commit().await?;

        Ok(())
    }
}

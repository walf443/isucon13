#[cfg(test)]
mod count_by_livestream_id;
#[cfg(test)]
mod create;
#[cfg(test)]
mod delete_by_livestream_id_and_user_id;

use crate::tables::livestream_viewers_history::LivestreamViewersHistoryRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;

#[derive(Clone)]
pub struct LivestreamViewersHistoryRepositoryInfra {}

#[async_trait]
impl LivestreamViewersHistoryRepository for LivestreamViewersHistoryRepositoryInfra {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        history: &CreateLivestreamViewersHistory,
    ) -> isupipe_core::repos::Result<()> {
        let mut tx = conn.transaction().await?;

        LivestreamViewersHistoryRow::create()
            .user_id(&history.user_id)
            .livestream_id(&history.livestream_id)
            .created_at(history.created_at)
            .exec(&mut tx)
            .await?;

        tx.commit().await?;

        Ok(())
    }

    async fn count_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let count = LivestreamViewersHistoryRow::filter(
            LivestreamViewersHistoryRow::fields()
                .livestream_id()
                .eq(livestream_id),
        )
        .count()
        .exec(conn)
        .await?;

        Ok(count as i64)
    }

    async fn delete_by_livestream_id_and_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<()> {
        let mut tx = conn.transaction().await?;

        LivestreamViewersHistoryRow::filter(
            LivestreamViewersHistoryRow::fields().user_id().eq(user_id),
        )
        .filter(
            LivestreamViewersHistoryRow::fields()
                .livestream_id()
                .eq(livestream_id),
        )
        .delete()
        .exec(&mut tx)
        .await?;

        tx.commit().await?;

        Ok(())
    }
}

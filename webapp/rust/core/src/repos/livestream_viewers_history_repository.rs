use crate::db::DBConn;
use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamViewersHistoryRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
        created_at: i64,
    ) -> Result<()>;
    async fn count_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> Result<i64>;
    async fn delete_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> Result<()>;
}

pub trait HaveLivestreamViewersHistoryRepository {
    type Repo: Sync + LivestreamViewersHistoryRepository;

    fn livestream_viewers_history_repo(&self) -> &Self::Repo;
}

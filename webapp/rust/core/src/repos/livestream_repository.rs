use crate::db::DBConn;
use crate::models::livestream::{CreateLivestream, Livestream, LivestreamId};
use crate::models::user::UserId;
use crate::repos::Result;
use async_trait::async_trait;

#[cfg_attr(any(feature = "test", test), mockall::automock)]
#[async_trait]
pub trait LivestreamRepository {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        stream: &CreateLivestream,
    ) -> Result<LivestreamId>;
    async fn find_all<'c>(&self, conn: &'c mut DBConn<'c>) -> Result<Vec<Livestream>>;

    async fn find_all_order_by_id_desc<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
    ) -> Result<Vec<Livestream>>;
    async fn find_all_order_by_id_desc_limit<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        limit: i64,
    ) -> Result<Vec<Livestream>>;

    async fn find_all_by_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user_id: &UserId,
    ) -> Result<Vec<Livestream>>;
    async fn find<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        id: &LivestreamId,
    ) -> Result<Option<Livestream>>;

    async fn exist_by_id_and_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        id: &LivestreamId,
        user_id: &UserId,
    ) -> Result<bool>;
}

pub trait HaveLivestreamRepository {
    type Repo: Sync + LivestreamRepository;

    fn livestream_repo(&self) -> &Self::Repo;
}

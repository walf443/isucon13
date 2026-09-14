use crate::db::DBConn;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_comment::{
    CreateLivestreamComment, LivestreamComment, LivestreamCommentId,
};
use crate::models::user::UserId;
use crate::repos::Result;
use async_trait::async_trait;

#[cfg_attr(any(feature = "test", test), mockall::automock)]
#[async_trait]
pub trait LivestreamCommentRepository {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        comment: &CreateLivestreamComment,
    ) -> Result<LivestreamCommentId>;

    async fn remove_if_match_ng_word<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        comment: &LivestreamComment,
        ng_word: &str,
    ) -> Result<()>;

    async fn find<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        comment_id: &LivestreamCommentId,
    ) -> Result<Option<LivestreamComment>>;

    async fn find_all<'c>(&self, conn: &'c mut DBConn<'c>) -> Result<Vec<LivestreamComment>>;

    async fn find_all_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<LivestreamComment>>;

    async fn find_all_by_livestream_id_order_by_created_at<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<LivestreamComment>>;

    async fn find_all_by_livestream_id_order_by_created_at_limit<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
        limit: i64,
    ) -> Result<Vec<LivestreamComment>>;

    async fn get_sum_tip<'c>(&self, conn: &'c mut DBConn<'c>) -> Result<i64>;

    async fn get_sum_tip_of_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> Result<i64>;

    async fn get_max_tip_of_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> Result<i64>;

    async fn get_sum_tip_of_livestream_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user_id: &UserId,
    ) -> Result<i64>;
}

pub trait HaveLivestreamCommentRepository {
    type Repo: Sync + LivestreamCommentRepository;
    fn livestream_comment_repo(&self) -> &Self::Repo;
}

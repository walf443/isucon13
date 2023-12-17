use crate::db::DBConn;
use crate::models::livestream_comment::LivestreamCommentModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamCommentRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        livestream_id: i64,
        comment: &str,
        tip: i64,
        created_at: i64,
    ) -> Result<i64>;

    async fn remove_if_match_ng_word(
        &self,
        conn: &mut DBConn,
        comment: &LivestreamCommentModel,
        ng_word: &str,
    ) -> Result<()>;

    async fn find(
        &self,
        conn: &mut DBConn,
        comment_id: i64,
    ) -> Result<Option<LivestreamCommentModel>>;

    async fn find_all(&self, conn: &mut DBConn) -> Result<Vec<LivestreamCommentModel>>;

    async fn find_all_by_livestream_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<LivestreamCommentModel>>;

    async fn find_all_by_livestream_id_order_by_created_at_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        limit: i64,
    ) -> Result<Vec<LivestreamCommentModel>>;

    async fn get_sum_of_tips(&self, conn: &mut DBConn) -> Result<i64>;

    async fn get_sum_of_tips_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<i64>;
}

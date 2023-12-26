use crate::db::DBConn;
use crate::models::livestream_comment::LivestreamComment;
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
        comment: &LivestreamComment,
        ng_word: &str,
    ) -> Result<()>;

    async fn find(&self, conn: &mut DBConn, comment_id: i64) -> Result<Option<LivestreamComment>>;

    async fn find_all(&self, conn: &mut DBConn) -> Result<Vec<LivestreamComment>>;

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<LivestreamComment>>;

    async fn find_all_by_livestream_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<LivestreamComment>>;

    async fn find_all_by_livestream_id_order_by_created_at_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        limit: i64,
    ) -> Result<Vec<LivestreamComment>>;

    async fn get_sum_tip(&self, conn: &mut DBConn) -> Result<i64>;

    async fn get_sum_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<i64>;

    async fn get_max_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<i64>;

    async fn get_sum_tip_of_livestream_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> Result<i64>;
}

pub trait HaveLivestreamCommentRepository {
    type Repo: Sync + LivestreamCommentRepository;
    fn livestream_comment_repo(&self) -> &Self::Repo;
}

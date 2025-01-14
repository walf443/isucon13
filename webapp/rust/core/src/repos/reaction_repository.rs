use crate::db::DBConn;
use crate::models::livestream::LivestreamId;
use crate::models::reaction::{CreateReaction, Reaction, ReactionId};
use crate::models::user::{UserId, UserName};
use crate::repos::Result;
use async_trait::async_trait;

#[cfg_attr(any(feature = "test", test), mockall::automock)]
#[async_trait]
pub trait ReactionRepository {
    async fn create(&self, conn: &mut DBConn, reaction: &CreateReaction) -> Result<ReactionId>;

    async fn count_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> Result<i64>;

    async fn most_favorite_emoji_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &UserName,
    ) -> Result<String>;

    async fn count_by_livestream_user_id(
        &self,
        conn: &mut DBConn,
        livestream_user_id: &UserId,
    ) -> Result<i64>;

    async fn count_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &UserName,
    ) -> Result<i64>;

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<Reaction>>;
    async fn find_all_by_livestream_id_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        limit: i64,
    ) -> Result<Vec<Reaction>>;
}

pub trait HaveReactionRepository {
    type Repo: Sync + ReactionRepository;

    fn reaction_repo(&self) -> &Self::Repo;
}

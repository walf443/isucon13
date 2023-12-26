use crate::db::DBConn;
use crate::models::reaction::{Reaction, ReactionId};
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait ReactionRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        livestream_id: i64,
        emoji_name: &str,
        created_at: i64,
    ) -> Result<ReactionId>;

    async fn count_by_livestream_id(&self, conn: &mut DBConn, livestream_id: i64) -> Result<i64>;

    async fn most_favorite_emoji_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &str,
    ) -> Result<String>;

    async fn count_by_livestream_user_id(
        &self,
        conn: &mut DBConn,
        livestream_user_id: i64,
    ) -> Result<i64>;

    async fn count_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &str,
    ) -> Result<i64>;

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<Reaction>>;
    async fn find_all_by_livestream_id_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        limit: i64,
    ) -> Result<Vec<Reaction>>;
}

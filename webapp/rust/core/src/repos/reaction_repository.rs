use crate::db::DBConn;
use crate::models::reaction::ReactionModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait ReactionRepository {
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<ReactionModel>>;
    async fn find_all_by_livestream_id_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        limit: i64,
    ) -> Result<Vec<ReactionModel>>;
}

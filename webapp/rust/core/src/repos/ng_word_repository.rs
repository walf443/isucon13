use crate::db::DBConn;
use crate::models::livestream::LivestreamId;
use crate::models::ng_word::{NgWord, NgWordId};
use crate::models::user::UserId;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait NgWordRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
        livestream_id: &LivestreamId,
        word: &str,
        created_at: i64,
    ) -> Result<NgWordId>;
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<NgWord>>;

    async fn find_all_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> Result<Vec<NgWord>>;

    async fn find_all_by_livestream_id_and_user_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> Result<Vec<NgWord>>;
}

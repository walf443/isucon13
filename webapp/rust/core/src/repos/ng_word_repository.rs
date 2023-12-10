use crate::db::DBConn;
use crate::models::ng_word::NgWord;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait NgWordRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        livestream_id: i64,
        word: &str,
        created_at: i64,
    ) -> Result<i64>;
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<NgWord>>;

    async fn find_all_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        user_id: i64,
    ) -> Result<Vec<NgWord>>;

    async fn find_all_by_livestream_id_and_user_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        user_id: i64,
    ) -> Result<Vec<NgWord>>;
}

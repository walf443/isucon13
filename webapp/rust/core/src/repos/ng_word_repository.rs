use crate::db::DBConn;
use crate::models::ng_word::NgWord;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait NgWordRepository {
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<NgWord>>;
}

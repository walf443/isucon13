use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::ng_word::NgWord;
use isupipe_core::repos::ng_word_repository::NgWordRepository;

pub struct NgWordRepositoryInfra {}

#[async_trait]
impl NgWordRepository for NgWordRepositoryInfra {
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> isupipe_core::repos::Result<Vec<NgWord>> {
        let ng_words = sqlx::query_as("SELECT * FROM ng_words WHERE livestream_id = ?")
            .bind(livestream_id)
            .fetch_all(conn)
            .await?;

        Ok(ng_words)
    }

    async fn find_all_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        user_id: i64,
    ) -> isupipe_core::repos::Result<Vec<NgWord>> {
        let ng_words: Vec<NgWord> =
            sqlx::query_as("SELECT id, user_id, livestream_id, word FROM ng_words WHERE user_id = ? AND livestream_id = ?")
                .bind(livestream_id)
                .bind(user_id)
                .fetch_all(conn)
                .await?;

        Ok(ng_words)
    }
}

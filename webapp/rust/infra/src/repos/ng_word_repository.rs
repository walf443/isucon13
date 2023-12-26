use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::ng_word::{NgWord, NgWordId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;

pub struct NgWordRepositoryInfra {}

#[async_trait]
impl NgWordRepository for NgWordRepositoryInfra {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
        livestream_id: &LivestreamId,
        word: &str,
        created_at: i64,
    ) -> isupipe_core::repos::Result<NgWordId> {
        let rs = sqlx::query(
            "INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (?, ?, ?, ?)",
        )
        .bind(user_id)
        .bind(livestream_id)
        .bind(word)
        .bind(created_at)
        .execute(conn)
        .await?;

        let word_id = rs.last_insert_id() as i64;

        Ok(NgWordId::new(word_id))
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
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
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Vec<NgWord>> {
        let ng_words: Vec<NgWord> =
            sqlx::query_as("SELECT id, user_id, livestream_id, word FROM ng_words WHERE user_id = ? AND livestream_id = ?")
                .bind(livestream_id)
                .bind(user_id)
                .fetch_all(conn)
                .await?;

        Ok(ng_words)
    }

    async fn find_all_by_livestream_id_and_user_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Vec<NgWord>> {
        let ng_words: Vec<NgWord> = sqlx::query_as(
            "SELECT * FROM ng_words WHERE user_id = ? AND livestream_id = ? ORDER BY created_at DESC",
        )
            .bind(user_id)
            .bind(livestream_id)
            .fetch_all(conn)
            .await?;

        Ok(ng_words)
    }
}

#[cfg(test)]
mod find_all_by_livestream_id;
#[cfg(test)]
mod find_all_by_livestream_id_and_user_id;
#[cfg(test)]
mod find_all_by_livestream_id_and_user_id_order_by_created_at;

use crate::sqipe_support::bind_sqipe_values;
use crate::tables::ng_word::TABLE_NG_WORDS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::ng_word::{CreateNgWord, NgWord, NgWordId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;
use sqipe_mysql::sqipe;

#[derive(Clone)]
pub struct NgWordRepositoryInfra {}

#[async_trait]
impl NgWordRepository for NgWordRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        ng_word: &CreateNgWord,
    ) -> isupipe_core::repos::Result<NgWordId> {
        let rs = sqlx::query(
            "INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (?, ?, ?, ?)",
        )
        .bind(&ng_word.user_id)
        .bind(&ng_word.livestream_id)
        .bind(&ng_word.word)
        .bind(ng_word.created_at)
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
        let t = &TABLE_NG_WORDS;
        let mut q = sqipe(t.table_name());
        q.and_where(("livestream_id", *livestream_id.inner()));
        let (sql, binds) = q.to_sql();
        let ng_words = bind_sqipe_values!(sqlx::query_as::<_, NgWord>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(ng_words)
    }

    async fn count_by_ng_word_in_comment(
        &self,
        conn: &mut DBConn,
        ng_word: &str,
        comment: &str,
    ) -> isupipe_core::repos::Result<i64> {
        // Complex CONCAT/LIKE subquery - not supported by squipe
        let query = r#"
        SELECT COUNT(*)
        FROM
        (SELECT ? AS text) AS texts
        INNER JOIN
        (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
        ON texts.text LIKE patterns.pattern;
        "#;
        let hit_spam: i64 = sqlx::query_scalar(query)
            .bind(comment)
            .bind(ng_word)
            .fetch_one(conn)
            .await?;

        Ok(hit_spam)
    }

    async fn find_all_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Vec<NgWord>> {
        let t = &TABLE_NG_WORDS;
        let mut q = sqipe(t.table_name());
        q.select(&["id", "user_id", "livestream_id", "word"]);
        q.and_where(("user_id", *user_id.inner()));
        q.and_where(("livestream_id", *livestream_id.inner()));
        let (sql, binds) = q.to_sql();
        let ng_words = bind_sqipe_values!(sqlx::query_as::<_, NgWord>(&sql), binds)
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
        let t = &TABLE_NG_WORDS;
        let mut q = sqipe(t.table_name());
        q.and_where(("user_id", *user_id.inner()));
        q.and_where(("livestream_id", *livestream_id.inner()));
        q.order_by(t.created_at().desc());
        let (sql, binds) = q.to_sql();
        let ng_words = bind_sqipe_values!(sqlx::query_as::<_, NgWord>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(ng_words)
    }
}

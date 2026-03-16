#[cfg(test)]
mod create;
#[cfg(test)]
mod find_all_by_livestream_id;
#[cfg(test)]
mod find_all_by_livestream_id_and_user_id;
#[cfg(test)]
mod find_all_by_livestream_id_and_user_id_order_by_created_at;

use crate::qbey_support::bind_qbey_values;
use crate::tables::ng_word::TABLE_NG_WORDS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::ng_word::{CreateNgWord, NgWord, NgWordId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;
use qbey_mysql::qbey;

struct InsertNgWord<'a>(&'a CreateNgWord);

impl qbey::ToInsertRow<qbey::Value> for InsertNgWord<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("user_id", (*self.0.user_id.inner()).into()),
            ("livestream_id", (*self.0.livestream_id.inner()).into()),
            ("word", self.0.word.as_str().into()),
            ("created_at", self.0.created_at.into()),
        ]
    }
}

#[derive(Clone)]
pub struct NgWordRepositoryInfra {}

#[async_trait]
impl NgWordRepository for NgWordRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        ng_word: &CreateNgWord,
    ) -> isupipe_core::repos::Result<NgWordId> {
        let t = &TABLE_NG_WORDS;
        let mut ins = qbey(t.table()).into_insert();
        ins.add_value(&InsertNgWord(ng_word));
        let (sql, binds) = ins.to_sql();
        let rs = bind_qbey_values!(sqlx::query(&sql), binds)
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
        let mut q = qbey(t.table());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        let (sql, binds) = q.to_sql();
        let ng_words = bind_qbey_values!(sqlx::query_as::<_, NgWord>(&sql), binds)
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
        // Complex CONCAT/LIKE subquery - not supported by qbey
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
        let mut q = qbey(t.table());
        q.and_where(t.user_id().eq(*user_id.inner()));
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        q.select(&[t.id(), t.user_id(), t.livestream_id(), t.word()]);
        let (sql, binds) = q.to_sql();
        let ng_words = bind_qbey_values!(sqlx::query_as::<_, NgWord>(&sql), binds)
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
        let mut q = qbey(t.table());
        q.and_where(t.user_id().eq(*user_id.inner()));
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        q.order_by(t.created_at().desc());
        let (sql, binds) = q.to_sql();
        let ng_words = bind_qbey_values!(sqlx::query_as::<_, NgWord>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(ng_words)
    }
}

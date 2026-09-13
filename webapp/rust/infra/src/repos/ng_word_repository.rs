#[cfg(test)]
mod create;
#[cfg(test)]
mod find_all_by_livestream_id;
#[cfg(test)]
mod find_all_by_livestream_id_and_user_id;
#[cfg(test)]
mod find_all_by_livestream_id_and_user_id_order_by_created_at;

use crate::sql_support::scalar_i64;
use crate::tables::ng_word::NgWordRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::ng_word::{CreateNgWord, NgWord, NgWordId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;

#[derive(Clone)]
pub struct NgWordRepositoryInfra {}

#[async_trait]
impl NgWordRepository for NgWordRepositoryInfra {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        ng_word: &CreateNgWord,
    ) -> isupipe_core::repos::Result<NgWordId> {
        let row = NgWordRow::create()
            .user_id(&ng_word.user_id)
            .livestream_id(&ng_word.livestream_id)
            .word(&ng_word.word)
            .created_at(ng_word.created_at)
            .exec(conn)
            .await?;

        Ok(row.id)
    }

    async fn find_all_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<NgWord>> {
        let rows = NgWordRow::filter(NgWordRow::fields().livestream_id().eq(livestream_id))
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn count_by_ng_word_in_comment<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        ng_word: &str,
        comment: &str,
    ) -> isupipe_core::repos::Result<i64> {
        // Complex CONCAT/LIKE subquery - not supported by toasty's query builder
        let query = r#"
        SELECT COUNT(*)
        FROM
        (SELECT ? AS text) AS texts
        INNER JOIN
        (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
        ON texts.text LIKE patterns.pattern
        "#;
        let rows = toasty::sql::query(query)
            .bind(comment)
            .bind(ng_word)
            .exec(conn)
            .await?;

        Ok(scalar_i64(rows)?)
    }

    async fn find_all_by_livestream_id_and_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Vec<NgWord>> {
        let rows = NgWordRow::filter(NgWordRow::fields().user_id().eq(user_id))
            .filter(NgWordRow::fields().livestream_id().eq(livestream_id))
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_by_livestream_id_and_user_id_order_by_created_at<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Vec<NgWord>> {
        let rows = NgWordRow::filter(NgWordRow::fields().user_id().eq(user_id))
            .filter(NgWordRow::fields().livestream_id().eq(livestream_id))
            .order_by(NgWordRow::fields().created_at().desc())
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }
}

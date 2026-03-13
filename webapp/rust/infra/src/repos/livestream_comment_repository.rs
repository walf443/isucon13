#[cfg(test)]
mod get_max_tip_of_livestream_id;
#[cfg(test)]
mod get_sum_tip;
#[cfg(test)]
mod get_sum_tip_of_livestream_id;
#[cfg(test)]
mod get_sum_tip_of_livestream_user_id;

use crate::sqipe_support::bind_sqipe_values;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{
    self, CreateLivestreamComment, LivestreamComment, LivestreamCommentId,
};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use isupipe_core::models::livestream::{self as livestream};
use isupipe_core::models::user::{self as user};
use sqipe::{aggregate, col, table};
use sqipe_mysql::sqipe;

#[derive(Clone)]
pub struct LivestreamCommentRepositoryInfra {}

#[async_trait]
impl LivestreamCommentRepository for LivestreamCommentRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        comment: &CreateLivestreamComment,
    ) -> isupipe_core::repos::Result<LivestreamCommentId> {
        let rs = sqlx::query(
            "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)",
        )
            .bind(&comment.user_id)
            .bind(&comment.livestream_id)
            .bind(&comment.comment)
            .bind(comment.tip)
            .bind(comment.created_at)
            .execute(conn)
            .await?;
        let comment_id = rs.last_insert_id() as i64;

        Ok(LivestreamCommentId::new(comment_id))
    }

    async fn remove_if_match_ng_word(
        &self,
        conn: &mut DBConn,
        comment: &LivestreamComment,
        ng_word: &str,
    ) -> isupipe_core::repos::Result<()> {
        // DELETE with complex subquery - not supported by squipe
        let query = r#"
        DELETE FROM livecomments
        WHERE
        id = ? AND
        livestream_id = ? AND
        (SELECT COUNT(*)
        FROM
        (SELECT ? AS text) AS texts
        INNER JOIN
        (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
        ON texts.text LIKE patterns.pattern) >= 1
        "#;
        sqlx::query(query)
            .bind(&comment.id)
            .bind(&comment.livestream_id)
            .bind(&comment.comment)
            .bind(ng_word)
            .execute(conn)
            .await?;

        Ok(())
    }

    async fn find(
        &self,
        conn: &mut DBConn,
        comment_id: &LivestreamCommentId,
    ) -> isupipe_core::repos::Result<Option<LivestreamComment>> {
        let mut q = sqipe(livestream_comment::TABLE_NAME);
        q.and_where(("id", *comment_id.inner()));
        let (sql, binds) = q.to_sql();
        let comment = bind_sqipe_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(comment)
    }

    async fn find_all(
        &self,
        conn: &mut DBConn,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let q = sqipe(livestream_comment::TABLE_NAME);
        let (sql, binds) = q.to_sql();
        let livecomments =
            bind_sqipe_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(livecomments)
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let mut q = sqipe(livestream_comment::TABLE_NAME);
        q.and_where(("livestream_id", *livestream_id.inner()));
        let (sql, binds) = q.to_sql();
        let comments =
            bind_sqipe_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(comments)
    }

    async fn find_all_by_livestream_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let mut q = sqipe(livestream_comment::TABLE_NAME);
        q.and_where(("livestream_id", *livestream_id.inner()));
        q.order_by(col("created_at").desc());
        let (sql, binds) = q.to_sql();
        let comments =
            bind_sqipe_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(comments)
    }

    async fn find_all_by_livestream_id_order_by_created_at_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let mut q = sqipe(livestream_comment::TABLE_NAME);
        q.and_where(("livestream_id", *livestream_id.inner()));
        q.order_by(col("created_at").desc());
        q.limit(limit as u64);
        let (sql, binds) = q.to_sql();
        let comments =
            bind_sqipe_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(comments)
    }

    async fn get_sum_tip(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<i64> {
        let mut q = sqipe(livestream_comment::TABLE_NAME);
        q.aggregate(&[aggregate::expr("CAST(IFNULL(SUM(tip), 0) AS SIGNED)")]);
        let (sql, binds) = q.to_sql();
        let total_tip = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(total_tip)
    }

    async fn get_sum_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let mut q = sqipe(livestream::TABLE_NAME);
        q.as_("l");
        q.join(
            livestream_comment::TABLE_NAME,
            table("l").col("id").eq_col("livestream_id"),
        );
        q.and_where(table("l").col("id").eq(*livestream_id.inner()));
        q.aggregate(&[aggregate::expr("CAST(IFNULL(SUM(livecomments.tip), 0) AS SIGNED)")]);
        let (sql, binds) = q.to_sql();
        let total_tips = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(total_tips)
    }

    async fn get_max_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let mut q = sqipe(livestream::TABLE_NAME);
        q.as_("l");
        q.join(
            livestream_comment::TABLE_NAME,
            table("l").col("id").eq_col("livestream_id"),
        );
        q.and_where(table("l").col("id").eq(*livestream_id.inner()));
        q.aggregate(&[aggregate::expr("CAST(IFNULL(MAX(livecomments.tip), 0) AS SIGNED)")]);

        let (sql, binds) = q.to_sql();
        let max_tip = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(max_tip)
    }

    async fn get_sum_tip_of_livestream_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<i64> {
        let mut q = sqipe(user::TABLE_NAME);
        q.as_("u");
        q.join(
            livestream::TABLE_NAME,
            table("u").col("id").eq_col("user_id"),
        );
        q.join(
            livestream_comment::TABLE_NAME,
            table(livestream::TABLE_NAME).col("id").eq_col("livestream_id"),
        );
        q.and_where(table("u").col("id").eq(*user_id.inner()));
        q.aggregate(&[aggregate::expr("CAST(IFNULL(SUM(livecomments.tip), 0) AS SIGNED)")]);

        let (sql, binds) = q.to_sql();
        let tips = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(tips)
    }
}

use crate::sqipe_support::bind_sqipe_values;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{
    self, CreateLivestreamComment, LivestreamComment, LivestreamCommentId,
};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use sqipe::col;
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

    // IFNULL(SUM(...), 0) - not supported by squipe
    async fn get_sum_tip(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<i64> {
        let total_tip = sqlx::query_scalar("SELECT IFNULL(SUM(tip), 0) FROM livecomments")
            .fetch_one(conn)
            .await?;

        Ok(total_tip)
    }

    // IFNULL(SUM(...), 0) with JOIN - not supported by squipe
    async fn get_sum_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let total_tips = sqlx::query_scalar("SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?")
            .bind(livestream_id)
            .fetch_one(conn)
            .await?;

        Ok(total_tips)
    }

    // IFNULL(MAX(...), 0) with JOIN - not supported by squipe
    async fn get_max_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let max_tip = sqlx::query_scalar("SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?")
            .bind(livestream_id)
            .fetch_one(conn)
            .await?;

        Ok(max_tip)
    }

    // IFNULL(SUM(...), 0) with 2 JOINs - not supported by squipe
    async fn get_sum_tip_of_livestream_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<i64> {
        let query = r#"
        SELECT IFNULL(SUM(l2.tip), 0) FROM users u
        INNER JOIN livestreams l ON l.user_id = u.id
        INNER JOIN livecomments l2 ON l2.livestream_id = l.id
        WHERE u.id = ?
        "#;
        let tips = sqlx::query_scalar(query)
            .bind(user_id)
            .fetch_one(conn)
            .await?;

        Ok(tips)
    }
}

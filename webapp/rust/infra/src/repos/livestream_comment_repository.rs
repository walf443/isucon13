use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{LivestreamComment, LivestreamCommentId};
use isupipe_core::models::mysql_decimal::MysqlDecimal;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

pub struct LivestreamCommentRepositoryInfra {}

#[async_trait]
impl LivestreamCommentRepository for LivestreamCommentRepositoryInfra {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
        livestream_id: &LivestreamId,
        comment: &str,
        tip: i64,
        created_at: i64,
    ) -> isupipe_core::repos::Result<LivestreamCommentId> {
        let rs = sqlx::query(
            "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)",
        )
            .bind(user_id)
            .bind(livestream_id)
            .bind(&comment)
            .bind(tip)
            .bind(created_at)
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
            .bind(&ng_word)
            .execute(conn)
            .await?;

        Ok(())
    }

    async fn find(
        &self,
        conn: &mut DBConn,
        comment_id: &LivestreamCommentId,
    ) -> isupipe_core::repos::Result<Option<LivestreamComment>> {
        let comment = sqlx::query_as("SELECT * FROM livecomments WHERE id = ?")
            .bind(comment_id)
            .fetch_optional(conn)
            .await?;

        Ok(comment)
    }

    async fn find_all(
        &self,
        conn: &mut DBConn,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let livecomments: Vec<LivestreamComment> = sqlx::query_as("SELECT * FROM livecomments")
            .fetch_all(conn)
            .await?;

        Ok(livecomments)
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let comments: Vec<LivestreamComment> =
            sqlx::query_as("SELECT * FROM livecomments WHERE livestream_id = ?")
                .bind(livestream_id)
                .fetch_all(conn)
                .await?;

        Ok(comments)
    }

    async fn find_all_by_livestream_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let query = "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC"
            .to_owned();

        let comments: Vec<LivestreamComment> = sqlx::query_as(&query)
            .bind(livestream_id)
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
        let query =
            "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?"
                .to_owned();

        let comments: Vec<LivestreamComment> = sqlx::query_as(&query)
            .bind(livestream_id)
            .bind(limit)
            .fetch_all(conn)
            .await?;

        Ok(comments)
    }

    async fn get_sum_tip(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<i64> {
        let MysqlDecimal(total_tip) =
            sqlx::query_scalar("SELECT IFNULL(SUM(tip), 0) FROM livecomments")
                .fetch_one(conn)
                .await?;

        Ok(total_tip)
    }

    async fn get_sum_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let MysqlDecimal(total_tips) = sqlx::query_scalar("SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?")
            .bind(livestream_id)
            .fetch_one(conn)
            .await?;

        Ok(total_tips)
    }

    async fn get_max_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let MysqlDecimal(max_tip) = sqlx::query_scalar("SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?")
            .bind(livestream_id)
            .fetch_one(conn)
            .await?;

        Ok(max_tip)
    }

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
        let MysqlDecimal(tips) = sqlx::query_scalar(query)
            .bind(user_id)
            .fetch_one(conn)
            .await?;

        Ok(tips)
    }
}

use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream_comment::LivestreamCommentModel;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

pub struct LivestreamCommentRepositoryInfra {}

#[async_trait]
impl LivestreamCommentRepository for LivestreamCommentRepositoryInfra {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        livestream_id: i64,
        comment: &str,
        tip: i64,
        created_at: i64,
    ) -> isupipe_core::repos::Result<i64> {
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

        Ok(comment_id)
    }

    async fn find(
        &self,
        conn: &mut DBConn,
        comment_id: i64,
    ) -> isupipe_core::repos::Result<Option<LivestreamCommentModel>> {
        let comment = sqlx::query_as("SELECT * FROM livecomments WHERE id = ?")
            .bind(comment_id)
            .fetch_optional(conn)
            .await?;

        Ok(comment)
    }

    async fn find_all(
        &self,
        conn: &mut DBConn,
    ) -> isupipe_core::repos::Result<Vec<LivestreamCommentModel>> {
        let livecomments: Vec<LivestreamCommentModel> =
            sqlx::query_as("SELECT * FROM livecomments")
                .fetch_all(conn)
                .await?;

        Ok(livecomments)
    }

    async fn find_all_by_livestream_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> isupipe_core::repos::Result<Vec<LivestreamCommentModel>> {
        let query = "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC"
            .to_owned();

        let comments: Vec<LivestreamCommentModel> = sqlx::query_as(&query)
            .bind(livestream_id)
            .fetch_all(conn)
            .await?;

        Ok(comments)
    }

    async fn find_all_by_livestream_id_order_by_created_at_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<LivestreamCommentModel>> {
        let query =
            "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?"
                .to_owned();

        let comments: Vec<LivestreamCommentModel> = sqlx::query_as(&query)
            .bind(livestream_id)
            .bind(limit)
            .fetch_all(conn)
            .await?;

        Ok(comments)
    }
}

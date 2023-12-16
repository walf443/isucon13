use async_trait::async_trait;
use isupipe_core::db::DBConn;
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
}

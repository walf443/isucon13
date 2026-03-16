use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamComment};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'test', 'Test', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't1', 'd1', 'http://p', 'http://t', 1000, 2000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let input = CreateLivestreamComment {
        user_id: UserId::new(1),
        livestream_id: LivestreamId::new(1),
        comment: "hello world".to_string(),
        tip: 500,
        created_at: 1234,
    };

    let repo = LivestreamCommentRepositoryInfra {};
    let comment_id = repo.create(&mut tx, &input).await.unwrap();

    let got: LivestreamComment = sqlx::query_as("SELECT * FROM livecomments WHERE id = ?")
        .bind(&comment_id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();

    assert_eq!(got.id, comment_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.comment, input.comment);
    assert_eq!(got.tip, input.tip);
    assert_eq!(got.created_at, input.created_at);
}

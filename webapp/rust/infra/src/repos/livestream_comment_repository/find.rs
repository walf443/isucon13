use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream_comment::LivestreamCommentId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn found_case() {
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

    sqlx::query(
        "INSERT INTO livecomments (id, user_id, livestream_id, comment, tip, created_at) VALUES (1, 1, 1, 'hello', 100, 500)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamCommentId::new(1)).await.unwrap();
    assert!(result.is_some());
    let comment = result.unwrap();
    assert_eq!(*comment.id.inner(), 1);
    assert_eq!(comment.comment, "hello");
    assert_eq!(comment.tip, 100);
}

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamCommentId::new(999)).await.unwrap();
    assert!(result.is_none());
}

use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn not_empty_case() {
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
        "INSERT INTO livecomments (id, user_id, livestream_id, comment, tip, created_at) VALUES (1, 1, 1, 'hello', 100, 500), (2, 1, 1, 'world', 0, 600)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(result.len(), 2);
}

use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn returns_limited_rows() {
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
        "INSERT INTO livecomments (id, user_id, livestream_id, comment, tip, created_at) VALUES (1, 1, 1, 'a', 0, 100), (2, 1, 1, 'b', 0, 300), (3, 1, 1, 'c', 0, 200)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_order_by_created_at_limit(&mut tx, &LivestreamId::new(1), 2)
        .await
        .unwrap();
    assert_eq!(result.len(), 2);
    assert_eq!(result[0].created_at, 300);
    assert_eq!(result[1].created_at, 200);
}

use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'test', 'Test', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't', 'd', 'http://p', 'http://t', 1000, 2000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let max = repo
        .get_max_tip_of_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(max, 0);
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
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't', 'd', 'http://p', 'http://t', 1000, 2000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (1, 1, 'a', 100, 1), (1, 1, 'b', 200, 2)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let max = repo
        .get_max_tip_of_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(max, 200);
}

use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserId;
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
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(total, 0);
}

#[tokio::test]
async fn not_empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'alice', 'Alice', 'pw', 'desc'), (2, 'bob', 'Bob', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't', 'd', 'http://p', 'http://t', 1000, 2000), (2, 1, 't2', 'd2', 'http://p2', 'http://t2', 3000, 4000), (3, 2, 't3', 'd3', 'http://p3', 'http://t3', 5000, 6000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (2, 1, 'a', 100, 1), (2, 1, 'b', 200, 2), (2, 2, 'c', 50, 3), (1, 3, 'd', 300, 4)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamCommentRepositoryInfra {};

    // alice owns livestream 1 and 2, tips: 100+200+50=350
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(total, 350);

    // bob owns livestream 3, tips: 300
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(total, 300);
}

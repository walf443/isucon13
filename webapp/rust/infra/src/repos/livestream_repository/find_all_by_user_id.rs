use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

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

    let repo = LivestreamRepositoryInfra {};
    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_user_id() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'alice', 'Alice', 'pw', 'desc'), (2, 'bob', 'Bob', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't1', 'd1', 'http://p', 'http://t', 1000, 2000), (2, 1, 't2', 'd2', 'http://p2', 'http://t2', 3000, 4000), (3, 2, 't3', 'd3', 'http://p3', 'http://t3', 5000, 6000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamRepositoryInfra {};

    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 2);

    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(result.len(), 1);
    assert_eq!(*result[0].id.inner(), 3);
}

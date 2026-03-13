use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

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
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 'title1', 'desc1', 'http://p', 'http://t', 1000, 2000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamId::new(1)).await.unwrap();
    assert!(result.is_some());
    let ls = result.unwrap();
    assert_eq!(*ls.id.inner(), 1);
    assert_eq!(ls.title, "title1");
    assert_eq!(ls.start_at, 1000);
    assert_eq!(ls.end_at, 2000);
}

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamId::new(999)).await.unwrap();
    assert!(result.is_none());
}

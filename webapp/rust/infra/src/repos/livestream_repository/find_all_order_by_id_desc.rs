use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find_all_order_by_id_desc(&mut tx).await.unwrap();
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
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't1', 'd1', 'http://p', 'http://t', 1000, 2000), (2, 1, 't2', 'd2', 'http://p2', 'http://t2', 3000, 4000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find_all_order_by_id_desc(&mut tx).await.unwrap();
    assert_eq!(result.len(), 2);
    assert_eq!(*result[0].id.inner(), 2);
    assert_eq!(*result[1].id.inner(), 1);
}

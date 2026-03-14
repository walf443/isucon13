use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;

#[tokio::test]
async fn deletes_matching_entry() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'alice', 'Alice', 'pw', 'desc'), (2, 'bob', 'Bob', 'pw', 'desc')",
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
        "INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES (1, 1, 100), (2, 1, 200)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamViewersHistoryRepositoryInfra {};
    repo.delete_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();

    // Only user 2's entry should remain
    let count = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(count, 1);
}

#[tokio::test]
async fn noop_when_no_match() {
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

    let repo = LivestreamViewersHistoryRepositoryInfra {};
    // Should not error even if no matching entry
    repo.delete_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();
}

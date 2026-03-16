use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;

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

    let input = CreateLivestreamViewersHistory {
        user_id: UserId::new(1),
        livestream_id: LivestreamId::new(1),
        created_at: 1500,
    };

    let repo = LivestreamViewersHistoryRepositoryInfra {};
    repo.create(&mut tx, &input).await.unwrap();

    let count: i64 =
        sqlx::query_scalar("SELECT COUNT(*) FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?")
            .bind(1_i64)
            .bind(1_i64)
            .fetch_one(&mut *tx)
            .await
            .unwrap();

    assert_eq!(count, 1);
}

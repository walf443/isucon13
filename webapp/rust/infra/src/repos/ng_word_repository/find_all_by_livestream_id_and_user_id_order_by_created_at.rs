use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;

#[tokio::test]
async fn returns_ordered_by_created_at_desc() {
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
        "INSERT INTO ng_words (id, user_id, livestream_id, word, created_at) VALUES (1, 1, 1, 'bad', 100), (2, 1, 1, 'ugly', 300), (3, 1, 1, 'evil', 200)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = NgWordRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_and_user_id_order_by_created_at(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 3);
    assert_eq!(result[0].created_at, 300);
    assert_eq!(result[1].created_at, 200);
    assert_eq!(result[2].created_at, 100);
}

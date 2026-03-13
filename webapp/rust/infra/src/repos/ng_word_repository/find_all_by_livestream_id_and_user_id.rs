use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;

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
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't1', 'd1', 'http://p', 'http://t', 1000, 2000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = NgWordRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_both() {
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
        "INSERT INTO ng_words (id, user_id, livestream_id, word, created_at) VALUES (1, 1, 1, 'bad', 100), (2, 2, 1, 'ugly', 200), (3, 1, 1, 'evil', 300)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = NgWordRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 2);

    let result = repo
        .find_all_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(result.len(), 1);
}

use crate::repos::reaction_repository::ReactionRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserName;
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[tokio::test]
async fn zero_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'alice', 'Alice', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .count_by_livestream_user_name(&mut tx, &UserName::new("alice".to_string()))
        .await
        .unwrap();
    assert_eq!(result, 0);
}

#[tokio::test]
async fn counts_correctly() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'alice', 'Alice', 'pw', 'desc'), (2, 'bob', 'Bob', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't1', 'd1', 'http://p', 'http://t', 1000, 2000), (2, 2, 't2', 'd2', 'http://p2', 'http://t2', 3000, 4000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO reactions (id, user_id, livestream_id, emoji_name, created_at) VALUES (1, 1, 1, '👍', 100), (2, 2, 1, '❤️', 200), (3, 1, 2, '😀', 300)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .count_by_livestream_user_name(&mut tx, &UserName::new("alice".to_string()))
        .await
        .unwrap();
    assert_eq!(result, 2);

    let result = repo
        .count_by_livestream_user_name(&mut tx, &UserName::new("bob".to_string()))
        .await
        .unwrap();
    assert_eq!(result, 1);
}

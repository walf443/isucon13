use crate::repos::reaction_repository::ReactionRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserName;
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[tokio::test]
async fn empty_case() {
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
        .most_favorite_emoji_by_livestream_user_name(&mut tx, &UserName::new("alice".to_string()))
        .await
        .unwrap();
    assert_eq!(result, "");
}

#[tokio::test]
async fn returns_most_frequent_emoji() {
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
        "INSERT INTO reactions (id, user_id, livestream_id, emoji_name, created_at) VALUES (1, 2, 1, 'like', 100), (2, 2, 1, 'like', 200), (3, 2, 1, 'heart', 300)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .most_favorite_emoji_by_livestream_user_name(&mut tx, &UserName::new("alice".to_string()))
        .await
        .unwrap();
    assert_eq!(result, "like");
}

#[tokio::test]
async fn tiebreak_by_emoji_name_desc() {
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

    // Both emojis have count=1, so tiebreak by emoji_name DESC -> "like" > "heart"
    sqlx::query(
        "INSERT INTO reactions (id, user_id, livestream_id, emoji_name, created_at) VALUES (1, 2, 1, 'heart', 100), (2, 2, 1, 'like', 200)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .most_favorite_emoji_by_livestream_user_name(&mut tx, &UserName::new("alice".to_string()))
        .await
        .unwrap();
    assert_eq!(result, "like");
}

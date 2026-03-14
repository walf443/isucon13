use crate::repos::icon_repository::IconRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::icon_repository::IconRepository;

#[tokio::test]
async fn deletes_icon() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'alice', 'Alice', 'pw', 'desc'), (2, 'bob', 'Bob', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO icons (id, user_id, image) VALUES (1, 1, X'89504E47'), (2, 2, X'89504E47')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = IconRepositoryInfra {};
    repo.delete_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();

    // user 1's icon should be gone
    let result = repo
        .find_image_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert!(result.is_none());

    // user 2's icon should remain
    let result = repo
        .find_image_by_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert!(result.is_some());
}

#[tokio::test]
async fn noop_when_no_icon() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'test', 'Test', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = IconRepositoryInfra {};
    // Should not error even if no icon exists
    repo.delete_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
}

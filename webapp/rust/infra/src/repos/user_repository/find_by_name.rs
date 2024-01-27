use crate::repos::user_repository::UserRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::user_repository::UserRepository;

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = UserRepositoryInfra {};
    let name: String = Faker.fake();
    let got_user = repo.find_by_name(&mut tx, &name).await.unwrap();
    assert!(got_user.is_none())
}

#[tokio::test]
async fn found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();

    let result = sqlx::query(
        "INSERT INTO users (name, description, display_name, password) VALUES (?, ?, ?, ?)",
    )
    .bind(&user.name)
    .bind(&user.description)
    .bind(&user.display_name)
    .bind(&user.password)
    .execute(&mut *tx)
    .await
    .unwrap();
    let user_id = result.last_insert_id() as i64;
    let user_id = UserId::new(user_id);

    let repo = UserRepositoryInfra {};
    let got_user = repo.find_by_name(&mut tx, &user.name).await.unwrap();
    assert!(got_user.is_some());
    assert_eq!(got_user.unwrap().id, user_id)
}

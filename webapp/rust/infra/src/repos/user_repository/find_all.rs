use crate::repos::user_repository::UserRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::user_repository::UserRepository;

#[tokio::test]
async fn empty_case() {
    let pool = get_db_pool().await.unwrap();

    let mut tx = pool.begin().await.unwrap();
    let repo = UserRepositoryInfra {};

    let users = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(users.len(), 0);
}

#[tokio::test]
async fn not_empty_case() {
    let pool = get_db_pool().await.unwrap();

    let mut tx = pool.begin().await.unwrap();
    let repo = UserRepositoryInfra {};

    let user_count = 2;
    let mut users: Vec<CreateUser> = Vec::with_capacity(user_count);
    for _ in 0..user_count {
        users.push(Faker.fake())
    }

    sqlx::query("INSERT INTO users (name, display_name, description, password) VALUES (?, ?, ?, ?), (?, ?, ?, ?)")
        .bind(&users[0].name)
        .bind(&users[0].display_name)
        .bind(&users[0].description)
        .bind(&users[0].password)
        .bind(&users[1].name)
        .bind(&users[1].display_name)
        .bind(&users[1].description)
        .bind(&users[1].password)
        .execute(&mut *tx).await.unwrap();

    let users = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(users.len(), user_count);
}

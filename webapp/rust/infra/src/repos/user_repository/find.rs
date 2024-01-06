use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::user_repository::UserRepository;
use crate::repos::user_repository::UserRepositoryInfra;

#[tokio::test]
async fn found_case() {
}

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user_id: UserId = Faker.fake();
    let repo = UserRepositoryInfra {};

    let user = repo.find(&mut *tx, &user_id).await.unwrap();
    assert!(user.is_none());
}

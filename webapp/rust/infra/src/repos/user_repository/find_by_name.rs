use crate::repos::user_repository::UserRepositoryInfra;
use crate::test_support::InsertUserSetup;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::user_repository::UserRepository;

#[tokio::test]
async fn not_found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = UserRepositoryInfra {};
    let name: String = Faker.fake();
    let got_user = repo.find_by_name(&mut tx, &name).await.unwrap();
    assert!(got_user.is_none())
}

#[tokio::test]
async fn found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();

    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }
    let user_id = UserId::new(1);

    let repo = UserRepositoryInfra {};
    let got_user = repo.find_by_name(&mut tx, &user.name).await.unwrap();
    assert!(got_user.is_some());
    assert_eq!(got_user.unwrap().id, user_id)
}

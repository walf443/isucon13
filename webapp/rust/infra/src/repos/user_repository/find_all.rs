use crate::repos::user_repository::UserRepositoryInfra;
use crate::test_support::InsertUserSetup;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::user_repository::UserRepository;

#[tokio::test]
async fn empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();
    let repo = UserRepositoryInfra {};

    let users = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(users.len(), 0);
}

#[tokio::test]
async fn not_empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();
    let repo = UserRepositoryInfra {};

    let user_count = 2;
    let mut users: Vec<CreateUser> = Vec::with_capacity(user_count);
    for _ in 0..user_count {
        users.push(Faker.fake())
    }

    {
        InsertUserSetup {
            id: 1,
            user: &users[0],
        }
        .insert(&mut tx)
        .await;
        InsertUserSetup {
            id: 2,
            user: &users[1],
        }
        .insert(&mut tx)
        .await;
    }

    let users = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(users.len(), user_count);
}

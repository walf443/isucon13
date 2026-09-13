use crate::repos::icon_repository::IconRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertIconSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::icon_repository::IconRepository;

#[tokio::test]
async fn deletes_icon() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user1: CreateUser = Faker.fake();
    let user2: CreateUser = Faker.fake();
    {
        InsertUserSetup {
            id: 1,
            user: &user1,
        }
        .insert(&mut tx)
        .await;
        InsertUserSetup {
            id: 2,
            user: &user2,
        }
        .insert(&mut tx)
        .await;
    }

    {
        InsertIconSetup {
            id: 1,
            user_id: 1,
            image: vec![0x89u8, 0x50, 0x4E, 0x47],
        }
        .insert(&mut tx)
        .await;
        InsertIconSetup {
            id: 2,
            user_id: 2,
            image: vec![0x89u8, 0x50, 0x4E, 0x47],
        }
        .insert(&mut tx)
        .await;
    }

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
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let repo = IconRepositoryInfra {};
    // Should not error even if no icon exists
    repo.delete_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
}

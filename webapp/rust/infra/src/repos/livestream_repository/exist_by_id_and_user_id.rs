use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_repository::LivestreamRepository;

#[tokio::test]
async fn exists_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let stream: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo
        .exist_by_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();
    assert!(result);
}

#[tokio::test]
async fn not_exists_wrong_user() {
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

    let stream: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo
        .exist_by_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(2))
        .await
        .unwrap();
    assert!(!result);
}

#[tokio::test]
async fn not_exists_wrong_id() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let stream: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo
        .exist_by_id_and_user_id(&mut tx, &LivestreamId::new(999), &UserId::new(1))
        .await
        .unwrap();
    assert!(!result);
}

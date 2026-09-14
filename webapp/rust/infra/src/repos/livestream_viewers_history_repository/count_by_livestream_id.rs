use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup, InsertViewersHistorySetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;

#[tokio::test]
async fn zero_case() {
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

    let repo = LivestreamViewersHistoryRepositoryInfra {};
    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 0);
}

#[tokio::test]
async fn counts_correctly() {
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

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream1,
        }
        .insert(&mut tx)
        .await;
        InsertLivestreamSetup {
            id: 2,
            user_id: 1,
            stream: &stream2,
        }
        .insert(&mut tx)
        .await;
    }

    {
        InsertViewersHistorySetup {
            user_id: 1,
            livestream_id: 1,
            created_at: 100,
        }
        .insert(&mut tx)
        .await;
        InsertViewersHistorySetup {
            user_id: 2,
            livestream_id: 1,
            created_at: 200,
        }
        .insert(&mut tx)
        .await;
        InsertViewersHistorySetup {
            user_id: 1,
            livestream_id: 2,
            created_at: 300,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamViewersHistoryRepositoryInfra {};
    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 2);

    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(2))
        .await
        .unwrap();
    assert_eq!(result, 1);
}

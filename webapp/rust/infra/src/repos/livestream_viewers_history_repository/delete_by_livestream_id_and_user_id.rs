use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup, InsertViewersHistorySetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;

#[tokio::test]
async fn deletes_matching_entry() {
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
    }

    let repo = LivestreamViewersHistoryRepositoryInfra {};
    repo.delete_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();

    // Only user 2's entry should remain
    let count = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(count, 1);
}

#[tokio::test]
async fn noop_when_no_match() {
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
    // Should not error even if no matching entry
    repo.delete_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();
}

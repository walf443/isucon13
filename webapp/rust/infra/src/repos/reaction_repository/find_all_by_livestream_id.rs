use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertReactionSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::reaction::CreateReaction;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[tokio::test]
async fn empty_case() {
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

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn returns_ordered_by_created_at_desc() {
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

    {
        let r1: CreateReaction = CreateReaction {
            created_at: 100,
            ..Faker.fake()
        };
        InsertReactionSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            reaction: &r1,
        }
        .insert(&mut tx)
        .await;
        let r2: CreateReaction = CreateReaction {
            created_at: 300,
            ..Faker.fake()
        };
        InsertReactionSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            reaction: &r2,
        }
        .insert(&mut tx)
        .await;
        let r3: CreateReaction = CreateReaction {
            created_at: 200,
            ..Faker.fake()
        };
        InsertReactionSetup {
            id: 3,
            user_id: 1,
            livestream_id: 1,
            reaction: &r3,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 3);
    assert_eq!(result[0].created_at, 300);
    assert_eq!(result[1].created_at, 200);
    assert_eq!(result[2].created_at, 100);
}

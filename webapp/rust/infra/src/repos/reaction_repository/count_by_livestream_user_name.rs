use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertReactionSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::reaction::CreateReaction;
use isupipe_core::models::user::{CreateUser, UserName};
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[tokio::test]
async fn zero_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .count_by_livestream_user_name(&mut tx, &UserName::new(user.name.clone()))
        .await
        .unwrap();
    assert_eq!(result, 0);
}

#[tokio::test]
async fn counts_correctly() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let alice: CreateUser = Faker.fake();
    let bob: CreateUser = Faker.fake();
    {
        InsertUserSetup {
            id: 1,
            user: &alice,
        }
        .insert(&mut tx)
        .await;
        InsertUserSetup { id: 2, user: &bob }.insert(&mut tx).await;
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
            user_id: 2,
            stream: &stream2,
        }
        .insert(&mut tx)
        .await;
    }

    {
        let r1: CreateReaction = Faker.fake();
        InsertReactionSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            reaction: &r1,
        }
        .insert(&mut tx)
        .await;
        let r2: CreateReaction = Faker.fake();
        InsertReactionSetup {
            id: 2,
            user_id: 2,
            livestream_id: 1,
            reaction: &r2,
        }
        .insert(&mut tx)
        .await;
        let r3: CreateReaction = Faker.fake();
        InsertReactionSetup {
            id: 3,
            user_id: 1,
            livestream_id: 2,
            reaction: &r3,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .count_by_livestream_user_name(&mut tx, &UserName::new(alice.name.clone()))
        .await
        .unwrap();
    assert_eq!(result, 2);

    let result = repo
        .count_by_livestream_user_name(&mut tx, &UserName::new(bob.name.clone()))
        .await
        .unwrap();
    assert_eq!(result, 1);
}

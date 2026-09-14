use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertReactionSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::reaction::CreateReaction;
use isupipe_core::models::user::{CreateUser, UserName};
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[tokio::test]
async fn empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .most_favorite_emoji_by_livestream_user_name(&mut tx, &UserName::new(user.name.clone()))
        .await
        .unwrap();
    assert_eq!(result, "");
}

#[tokio::test]
async fn returns_most_frequent_emoji() {
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
            emoji_name: "like".to_string(),
            created_at: 100,
            ..Faker.fake()
        };
        InsertReactionSetup {
            id: 1,
            user_id: 2,
            livestream_id: 1,
            reaction: &r1,
        }
        .insert(&mut tx)
        .await;
        let r2: CreateReaction = CreateReaction {
            emoji_name: "like".to_string(),
            created_at: 200,
            ..Faker.fake()
        };
        InsertReactionSetup {
            id: 2,
            user_id: 2,
            livestream_id: 1,
            reaction: &r2,
        }
        .insert(&mut tx)
        .await;
        let r3: CreateReaction = CreateReaction {
            emoji_name: "heart".to_string(),
            created_at: 300,
            ..Faker.fake()
        };
        InsertReactionSetup {
            id: 3,
            user_id: 2,
            livestream_id: 1,
            reaction: &r3,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .most_favorite_emoji_by_livestream_user_name(&mut tx, &UserName::new(alice.name.clone()))
        .await
        .unwrap();
    assert_eq!(result, "like");
}

#[tokio::test]
async fn tiebreak_by_emoji_name_desc() {
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

    // Both emojis have count=1, so tiebreak by emoji_name DESC -> "like" > "heart"
    {
        let r1: CreateReaction = CreateReaction {
            emoji_name: "heart".to_string(),
            created_at: 100,
            ..Faker.fake()
        };
        InsertReactionSetup {
            id: 1,
            user_id: 2,
            livestream_id: 1,
            reaction: &r1,
        }
        .insert(&mut tx)
        .await;
        let r2: CreateReaction = CreateReaction {
            emoji_name: "like".to_string(),
            created_at: 200,
            ..Faker.fake()
        };
        InsertReactionSetup {
            id: 2,
            user_id: 2,
            livestream_id: 1,
            reaction: &r2,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .most_favorite_emoji_by_livestream_user_name(&mut tx, &UserName::new(alice.name.clone()))
        .await
        .unwrap();
    assert_eq!(result, "like");
}

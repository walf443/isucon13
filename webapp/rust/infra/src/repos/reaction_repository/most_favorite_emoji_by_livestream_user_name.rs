use crate::qbey_support::bind_qbey_values;
use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::reaction::TABLE_REACTIONS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{InsertLivestreamSetup, InsertReactionSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::reaction::CreateReaction;
use isupipe_core::models::user::{CreateUser, UserName};
use isupipe_core::repos::reaction_repository::ReactionRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
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
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let alice: CreateUser = Faker.fake();
    let bob: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup {
            id: 1,
            user: &alice,
        });
        ins.add_value(&InsertUserSetup { id: 2, user: &bob });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    {
        let mut ins = qbey(TABLE_REACTIONS.table()).into_insert();
        let r1: CreateReaction = CreateReaction {
            emoji_name: "like".to_string(),
            created_at: 100,
            ..Faker.fake()
        };
        ins.add_value(&InsertReactionSetup {
            id: 1,
            user_id: 2,
            livestream_id: 1,
            reaction: &r1,
        });
        let r2: CreateReaction = CreateReaction {
            emoji_name: "like".to_string(),
            created_at: 200,
            ..Faker.fake()
        };
        ins.add_value(&InsertReactionSetup {
            id: 2,
            user_id: 2,
            livestream_id: 1,
            reaction: &r2,
        });
        let r3: CreateReaction = CreateReaction {
            emoji_name: "heart".to_string(),
            created_at: 300,
            ..Faker.fake()
        };
        ins.add_value(&InsertReactionSetup {
            id: 3,
            user_id: 2,
            livestream_id: 1,
            reaction: &r3,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
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
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let alice: CreateUser = Faker.fake();
    let bob: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup {
            id: 1,
            user: &alice,
        });
        ins.add_value(&InsertUserSetup { id: 2, user: &bob });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    // Both emojis have count=1, so tiebreak by emoji_name DESC -> "like" > "heart"
    {
        let mut ins = qbey(TABLE_REACTIONS.table()).into_insert();
        let r1: CreateReaction = CreateReaction {
            emoji_name: "heart".to_string(),
            created_at: 100,
            ..Faker.fake()
        };
        ins.add_value(&InsertReactionSetup {
            id: 1,
            user_id: 2,
            livestream_id: 1,
            reaction: &r1,
        });
        let r2: CreateReaction = CreateReaction {
            emoji_name: "like".to_string(),
            created_at: 200,
            ..Faker.fake()
        };
        ins.add_value(&InsertReactionSetup {
            id: 2,
            user_id: 2,
            livestream_id: 1,
            reaction: &r2,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .most_favorite_emoji_by_livestream_user_name(&mut tx, &UserName::new(alice.name.clone()))
        .await
        .unwrap();
    assert_eq!(result, "like");
}

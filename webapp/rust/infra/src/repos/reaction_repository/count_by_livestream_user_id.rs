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
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::reaction_repository::ReactionRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn zero_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .count_by_livestream_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 0);
}

#[tokio::test]
async fn counts_correctly() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let alice: CreateUser = Faker.fake();
    let bob: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &alice });
        ins.add_value(&InsertUserSetup { id: 2, user: &bob });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup { id: 1, user_id: 1, stream: &stream1 });
        ins.add_value(&InsertLivestreamSetup { id: 2, user_id: 2, stream: &stream2 });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_REACTIONS.table()).into_insert();
        let r1: CreateReaction = Faker.fake();
        ins.add_value(&InsertReactionSetup { id: 1, user_id: 1, livestream_id: 1, reaction: &r1 });
        let r2: CreateReaction = Faker.fake();
        ins.add_value(&InsertReactionSetup { id: 2, user_id: 2, livestream_id: 1, reaction: &r2 });
        let r3: CreateReaction = Faker.fake();
        ins.add_value(&InsertReactionSetup { id: 3, user_id: 1, livestream_id: 2, reaction: &r3 });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = ReactionRepositoryInfra {};
    // User 1 owns livestream 1, which has 2 reactions
    let result = repo
        .count_by_livestream_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 2);

    // User 2 owns livestream 2, which has 1 reaction
    let result = repo
        .count_by_livestream_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(result, 1);
}

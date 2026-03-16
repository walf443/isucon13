use crate::qbey_support::bind_qbey_values;
use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::reaction::TABLE_REACTIONS;
use crate::tables::user::TABLE_USERS;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::reaction::CreateReaction;
use isupipe_core::models::user::{CreateUser, UserName};
use isupipe_core::repos::reaction_repository::ReactionRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn zero_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("name", user.name.as_str().into()),
            ("display_name", user.display_name.as_str().into()),
            ("password", user.password.as_str().into()),
            ("description", user.description.as_str().into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
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
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let alice: CreateUser = Faker.fake();
    let bob: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("name", alice.name.as_str().into()),
            ("display_name", alice.display_name.as_str().into()),
            ("password", alice.password.as_str().into()),
            ("description", alice.description.as_str().into()),
        ]);
        ins.add_value(&[
            ("id", 2i64.into()),
            ("name", bob.name.as_str().into()),
            ("display_name", bob.display_name.as_str().into()),
            ("password", bob.password.as_str().into()),
            ("description", bob.description.as_str().into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("user_id", 1i64.into()),
            ("title", stream1.title.as_str().into()),
            ("description", stream1.description.as_str().into()),
            ("playlist_url", stream1.playlist_url.as_str().into()),
            ("thumbnail_url", stream1.thumbnail_url.as_str().into()),
            ("start_at", stream1.start_at.into()),
            ("end_at", stream1.end_at.into()),
        ]);
        ins.add_value(&[
            ("id", 2i64.into()),
            ("user_id", 2i64.into()),
            ("title", stream2.title.as_str().into()),
            ("description", stream2.description.as_str().into()),
            ("playlist_url", stream2.playlist_url.as_str().into()),
            ("thumbnail_url", stream2.thumbnail_url.as_str().into()),
            ("start_at", stream2.start_at.into()),
            ("end_at", stream2.end_at.into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_REACTIONS.table()).into_insert();
        let r1: CreateReaction = Faker.fake();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("emoji_name", r1.emoji_name.as_str().into()), ("created_at", r1.created_at.into())]);
        let r2: CreateReaction = Faker.fake();
        ins.add_value(&[("id", 2i64.into()), ("user_id", 2i64.into()), ("livestream_id", 1i64.into()), ("emoji_name", r2.emoji_name.as_str().into()), ("created_at", r2.created_at.into())]);
        let r3: CreateReaction = Faker.fake();
        ins.add_value(&[("id", 3i64.into()), ("user_id", 1i64.into()), ("livestream_id", 2i64.into()), ("emoji_name", r3.emoji_name.as_str().into()), ("created_at", r3.created_at.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
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

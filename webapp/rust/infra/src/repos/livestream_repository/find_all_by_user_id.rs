use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn empty_case() {
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
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_user_id() {
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
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    let stream3: CreateLivestream = Faker.fake();
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
            ("user_id", 1i64.into()),
            ("title", stream2.title.as_str().into()),
            ("description", stream2.description.as_str().into()),
            ("playlist_url", stream2.playlist_url.as_str().into()),
            ("thumbnail_url", stream2.thumbnail_url.as_str().into()),
            ("start_at", stream2.start_at.into()),
            ("end_at", stream2.end_at.into()),
        ]);
        ins.add_value(&[
            ("id", 3i64.into()),
            ("user_id", 2i64.into()),
            ("title", stream3.title.as_str().into()),
            ("description", stream3.description.as_str().into()),
            ("playlist_url", stream3.playlist_url.as_str().into()),
            ("thumbnail_url", stream3.thumbnail_url.as_str().into()),
            ("start_at", stream3.start_at.into()),
            ("end_at", stream3.end_at.into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamRepositoryInfra {};

    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 2);

    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(result.len(), 1);
    assert_eq!(*result[0].id.inner(), 3);
}

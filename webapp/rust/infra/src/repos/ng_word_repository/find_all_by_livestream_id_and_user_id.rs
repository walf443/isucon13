use crate::qbey_support::bind_qbey_values;
use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::ng_word::TABLE_NG_WORDS;
use crate::tables::user::TABLE_USERS;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::ng_word::CreateNgWord;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::ng_word_repository::NgWordRepository;
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
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("user_id", 1i64.into()),
            ("title", stream.title.as_str().into()),
            ("description", stream.description.as_str().into()),
            ("playlist_url", stream.playlist_url.as_str().into()),
            ("thumbnail_url", stream.thumbnail_url.as_str().into()),
            ("start_at", stream.start_at.into()),
            ("end_at", stream.end_at.into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = NgWordRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_both() {
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

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("user_id", 1i64.into()),
            ("title", stream.title.as_str().into()),
            ("description", stream.description.as_str().into()),
            ("playlist_url", stream.playlist_url.as_str().into()),
            ("thumbnail_url", stream.thumbnail_url.as_str().into()),
            ("start_at", stream.start_at.into()),
            ("end_at", stream.end_at.into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_NG_WORDS.table()).into_insert();
        let w1: CreateNgWord = Faker.fake();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("word", w1.word.as_str().into()), ("created_at", w1.created_at.into())]);
        let w2: CreateNgWord = Faker.fake();
        ins.add_value(&[("id", 2i64.into()), ("user_id", 2i64.into()), ("livestream_id", 1i64.into()), ("word", w2.word.as_str().into()), ("created_at", w2.created_at.into())]);
        let w3: CreateNgWord = Faker.fake();
        ins.add_value(&[("id", 3i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("word", w3.word.as_str().into()), ("created_at", w3.created_at.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = NgWordRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 2);

    let result = repo
        .find_all_by_livestream_id_and_user_id(&mut tx, &LivestreamId::new(1), &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(result.len(), 1);
}

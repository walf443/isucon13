use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_tag::TABLE_LIVESTREAM_TAGS;
use crate::tables::tag::TABLE_TAGS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{
    InsertLivestreamSetup, InsertLivestreamTagSetup, InsertTagSetup, InsertUserSetup,
};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::tag::Tag;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;
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

    let repo = LivestreamTagRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_livestream_id() {
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

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream1,
        });
        ins.add_value(&InsertLivestreamSetup {
            id: 2,
            user_id: 1,
            stream: &stream2,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let tag1: Tag = Faker.fake();
    let tag2: Tag = Faker.fake();
    {
        let mut ins = qbey(TABLE_TAGS.table()).into_insert();
        ins.add_value(&InsertTagSetup { id: 1, tag: &tag1 });
        ins.add_value(&InsertTagSetup { id: 2, tag: &tag2 });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVESTREAM_TAGS.table()).into_insert();
        ins.add_value(&InsertLivestreamTagSetup {
            id: 1,
            livestream_id: 1,
            tag_id: 1,
        });
        ins.add_value(&InsertLivestreamTagSetup {
            id: 2,
            livestream_id: 1,
            tag_id: 2,
        });
        ins.add_value(&InsertLivestreamTagSetup {
            id: 3,
            livestream_id: 2,
            tag_id: 1,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamTagRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 2);

    let result = repo
        .find_all_by_livestream_id(&mut tx, &LivestreamId::new(2))
        .await
        .unwrap();
    assert_eq!(result.len(), 1);
}

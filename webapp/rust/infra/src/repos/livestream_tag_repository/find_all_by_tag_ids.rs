use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_tag::TABLE_LIVESTREAM_TAGS;
use crate::tables::tag::TABLE_TAGS;
use crate::tables::user::TABLE_USERS;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = LivestreamTagRepositoryInfra {};
    let result = repo
        .find_all_by_tag_ids(&mut tx, &[TagId::new(999)])
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_tag_ids() {
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
            ("user_id", 1i64.into()),
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

    let tag1: Tag = Faker.fake();
    let tag2: Tag = Faker.fake();
    let tag3: Tag = Faker.fake();
    {
        let mut ins = qbey(TABLE_TAGS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("name", tag1.name.inner().as_str().into())]);
        ins.add_value(&[("id", 2i64.into()), ("name", tag2.name.inner().as_str().into())]);
        ins.add_value(&[("id", 3i64.into()), ("name", tag3.name.inner().as_str().into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVESTREAM_TAGS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("livestream_id", 1i64.into()), ("tag_id", 1i64.into())]);
        ins.add_value(&[("id", 2i64.into()), ("livestream_id", 1i64.into()), ("tag_id", 2i64.into())]);
        ins.add_value(&[("id", 3i64.into()), ("livestream_id", 2i64.into()), ("tag_id", 1i64.into())]);
        ins.add_value(&[("id", 4i64.into()), ("livestream_id", 2i64.into()), ("tag_id", 3i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = LivestreamTagRepositoryInfra {};
    let result = repo
        .find_all_by_tag_ids(&mut tx, &[TagId::new(1)])
        .await
        .unwrap();
    assert_eq!(result.len(), 2);
    // Ordered by livestream_id DESC
    assert_eq!(*result[0].livestream_id.inner(), 2);
    assert_eq!(*result[1].livestream_id.inner(), 1);

    let result = repo
        .find_all_by_tag_ids(&mut tx, &[TagId::new(1), TagId::new(2)])
        .await
        .unwrap();
    assert_eq!(result.len(), 3);
}

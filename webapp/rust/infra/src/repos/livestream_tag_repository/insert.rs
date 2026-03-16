use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_tag::TABLE_LIVESTREAM_TAGS;
use crate::tables::tag::TABLE_TAGS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{InsertLivestreamSetup, InsertTagSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_tag::LivestreamTag;
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup { id: 1, user_id: 1, stream: &stream });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let tag: Tag = Faker.fake();
    {
        let mut ins = qbey(TABLE_TAGS.table()).into_insert();
        ins.add_value(&InsertTagSetup { id: 1, tag: &tag });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamTagRepositoryInfra {};
    repo.insert(&mut tx, &LivestreamId::new(1), &TagId::new(1))
        .await
        .unwrap();

    let t = &TABLE_LIVESTREAM_TAGS;
    let mut q = qbey(t.table());
    q.and_where(t.livestream_id().eq(1_i64));
    q.and_where(t.tag_id().eq(1_i64));
    let (sql, binds) = q.to_sql();
    let got: LivestreamTag =
        bind_qbey_values!(sqlx::query_as::<_, LivestreamTag>(&sql), binds)
            .fetch_one(&mut *tx)
            .await
            .unwrap();

    assert_eq!(*got.livestream_id.inner(), 1);
    assert_eq!(*got.tag_id.inner(), 1);
}

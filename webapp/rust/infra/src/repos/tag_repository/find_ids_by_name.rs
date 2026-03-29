use crate::qbey_support::bind_qbey_values;
use crate::repos::tag_repository::TagRepositoryInfra;
use crate::tables::tag::TABLE_TAGS;
use crate::test_support::InsertTagSetup;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::tag::Tag;
use isupipe_core::repos::tag_repository::TagRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let tag1: Tag = Faker.fake();

    let repo = TagRepositoryInfra {};
    let result = repo.find_ids_by_name(&mut tx, &tag1.name).await.unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let tag1: Tag = Faker.fake();

    {
        let mut ins = qbey(TABLE_TAGS.table()).into_insert();
        ins.add_value(&InsertTagSetup {
            id: *tag1.id.inner(),
            tag: &tag1,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = TagRepositoryInfra {};
    let result = repo.find_ids_by_name(&mut tx, &tag1.name).await.unwrap();
    assert_eq!(result.len(), 1);
    let got = result.first().unwrap();
    assert_eq!(got, &tag1.id)
}

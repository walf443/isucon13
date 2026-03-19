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
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tags = repo.find_all(&mut tx).await.unwrap();

    assert_eq!(tags.len(), 0)
}
#[tokio::test]
async fn exists_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let tag1: Tag = Faker.fake();
    let tag2: Tag = Faker.fake();

    {
        let mut ins = qbey(TABLE_TAGS.table()).into_insert();
        ins.add_value(&InsertTagSetup {
            id: *tag1.id.inner(),
            tag: &tag1,
        });
        ins.add_value(&InsertTagSetup {
            id: *tag2.id.inner(),
            tag: &tag2,
        });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = TagRepositoryInfra {};
    let tags = repo.find_all(&mut tx).await.unwrap();

    assert_eq!(tags.len(), 2);

    let tid = tag1.id.clone();
    let got = tags.iter().find(move |t| t.id == tid).unwrap();
    assert_eq!(got.name, tag1.name);

    let tid = tag2.id.clone();
    let got = tags.iter().find(move |t| t.id == tid).unwrap();
    assert_eq!(got.name, tag2.name);
}

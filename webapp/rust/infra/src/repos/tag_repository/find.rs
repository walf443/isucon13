use crate::qbey_support::bind_qbey_values;
use crate::repos::tag_repository::TagRepositoryInfra;
use crate::tables::tag::TABLE_TAGS;
use crate::test_support::InsertTagSetup;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::repos::tag_repository::TagRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tag_id: TagId = Faker.fake();
    let result = repo.find(&mut tx, &tag_id).await;
    assert!(result.is_err());
    let err_msg = format!("{}", result.unwrap_err());
    assert!(
        err_msg.contains("no rows returned"),
        "expected RowNotFound, got: {}",
        err_msg
    );
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tag: Tag = Faker.fake();

    {
        let mut ins = qbey(TABLE_TAGS.table()).into_insert();
        ins.add_value(&InsertTagSetup {
            id: *tag.id.inner(),
            tag: &tag,
        });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let got = repo.find(&mut tx, &tag.id).await.unwrap();
    assert_eq!(got.id, tag.id);
    assert_eq!(got.name, tag.name);
}

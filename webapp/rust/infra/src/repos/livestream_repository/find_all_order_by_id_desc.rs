use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

#[tokio::test]
async fn empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find_all_order_by_id_desc(&mut tx).await.unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn not_empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream1,
        }
        .insert(&mut tx)
        .await;
        InsertLivestreamSetup {
            id: 2,
            user_id: 1,
            stream: &stream2,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find_all_order_by_id_desc(&mut tx).await.unwrap();
    assert_eq!(result.len(), 2);
    assert_eq!(*result[0].id.inner(), 2);
    assert_eq!(*result[1].id.inner(), 1);
}

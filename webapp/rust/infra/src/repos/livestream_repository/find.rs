use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use crate::test_support::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

#[tokio::test]
async fn found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
                InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let stream: CreateLivestream = Faker.fake();
    {
                InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        }.insert(&mut tx).await;
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamId::new(1)).await.unwrap();
    assert!(result.is_some());
    let ls = result.unwrap();
    assert_eq!(*ls.id.inner(), 1);
    assert_eq!(ls.title, stream.title);
    assert_eq!(ls.start_at, stream.start_at);
    assert_eq!(ls.end_at, stream.end_at);
}

#[tokio::test]
async fn not_found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamId::new(999)).await.unwrap();
    assert!(result.is_none());
}

use crate::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{
    InsertLivestreamSetup, InsertLivestreamTagSetup, InsertTagSetup, InsertUserSetup,
};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;

#[tokio::test]
async fn empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = LivestreamTagRepositoryInfra {};
    let result = repo
        .find_all_by_tag_ids(&mut tx, &[TagId::new(999)])
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_tag_ids() {
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

    let tag1: Tag = Faker.fake();
    let tag2: Tag = Faker.fake();
    let tag3: Tag = Faker.fake();
    {
        InsertTagSetup { id: 1, tag: &tag1 }.insert(&mut tx).await;
        InsertTagSetup { id: 2, tag: &tag2 }.insert(&mut tx).await;
        InsertTagSetup { id: 3, tag: &tag3 }.insert(&mut tx).await;
    }

    {
        InsertLivestreamTagSetup {
            id: 1,
            livestream_id: 1,
            tag_id: 1,
        }
        .insert(&mut tx)
        .await;
        InsertLivestreamTagSetup {
            id: 2,
            livestream_id: 1,
            tag_id: 2,
        }
        .insert(&mut tx)
        .await;
        InsertLivestreamTagSetup {
            id: 3,
            livestream_id: 2,
            tag_id: 1,
        }
        .insert(&mut tx)
        .await;
        InsertLivestreamTagSetup {
            id: 4,
            livestream_id: 2,
            tag_id: 3,
        }
        .insert(&mut tx)
        .await;
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

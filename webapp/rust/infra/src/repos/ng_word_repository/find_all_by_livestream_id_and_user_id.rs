use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertNgWordSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::ng_word::CreateNgWord;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::ng_word_repository::NgWordRepository;

#[tokio::test]
async fn empty_case() {
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
        }
        .insert(&mut tx)
        .await;
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
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let alice: CreateUser = Faker.fake();
    let bob: CreateUser = Faker.fake();
    {
        InsertUserSetup {
            id: 1,
            user: &alice,
        }
        .insert(&mut tx)
        .await;
        InsertUserSetup { id: 2, user: &bob }.insert(&mut tx).await;
    }

    let stream: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        }
        .insert(&mut tx)
        .await;
    }

    {
        let w1: CreateNgWord = Faker.fake();
        InsertNgWordSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            ng_word: &w1,
        }
        .insert(&mut tx)
        .await;
        let w2: CreateNgWord = Faker.fake();
        InsertNgWordSetup {
            id: 2,
            user_id: 2,
            livestream_id: 1,
            ng_word: &w2,
        }
        .insert(&mut tx)
        .await;
        let w3: CreateNgWord = Faker.fake();
        InsertNgWordSetup {
            id: 3,
            user_id: 1,
            livestream_id: 1,
            ng_word: &w3,
        }
        .insert(&mut tx)
        .await;
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

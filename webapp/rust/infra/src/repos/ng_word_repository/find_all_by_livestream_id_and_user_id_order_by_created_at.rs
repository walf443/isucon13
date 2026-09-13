use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertNgWordSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::ng_word::CreateNgWord;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::ng_word_repository::NgWordRepository;

#[tokio::test]
async fn returns_ordered_by_created_at_desc() {
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

    {
        let mut w1: CreateNgWord = Faker.fake();
        w1.created_at = 100;
        InsertNgWordSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            ng_word: &w1,
        }
        .insert(&mut tx)
        .await;
        let mut w2: CreateNgWord = Faker.fake();
        w2.created_at = 300;
        InsertNgWordSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            ng_word: &w2,
        }
        .insert(&mut tx)
        .await;
        let mut w3: CreateNgWord = Faker.fake();
        w3.created_at = 200;
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
        .find_all_by_livestream_id_and_user_id_order_by_created_at(
            &mut tx,
            &LivestreamId::new(1),
            &UserId::new(1),
        )
        .await
        .unwrap();
    assert_eq!(result.len(), 3);
    assert_eq!(result[0].created_at, 300);
    assert_eq!(result[1].created_at, 200);
    assert_eq!(result[2].created_at, 100);
}

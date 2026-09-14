use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use crate::tables::ng_word::NgWordRow;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::ng_word::{CreateNgWord, NgWord};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::ng_word_repository::NgWordRepository;

#[tokio::test]
async fn success_case() {
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

    let mut input: CreateNgWord = Faker.fake();
    input.user_id = UserId::new(1);
    input.livestream_id = LivestreamId::new(1);

    let repo = NgWordRepositoryInfra {};
    let word_id = repo.create(&mut tx, &input).await.unwrap();

    let got: NgWord = NgWordRow::all()
        .filter(NgWordRow::fields().id().eq(&word_id))
        .one()
        .exec(&mut tx)
        .await
        .unwrap()
        .into();

    assert_eq!(got.id, word_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.word, input.word);
    assert_eq!(got.created_at, input.created_at);
}

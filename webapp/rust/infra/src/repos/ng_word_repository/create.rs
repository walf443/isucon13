use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::ng_word::{CreateNgWord, NgWord};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::ng_word_repository::NgWordRepository;
use qbey_mysql::qbey;
use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;

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

    let mut input: CreateNgWord = Faker.fake();
    input.user_id = UserId::new(1);
    input.livestream_id = LivestreamId::new(1);

    let repo = NgWordRepositoryInfra {};
    let word_id = repo.create(&mut tx, &input).await.unwrap();

    let got: NgWord = sqlx::query_as("SELECT * FROM ng_words WHERE id = ?")
        .bind(&word_id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();

    assert_eq!(got.id, word_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.word, input.word);
    assert_eq!(got.created_at, input.created_at);
}

use crate::qbey_support::bind_qbey_values;
use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::ng_word::{CreateNgWord, NgWord};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::ng_word_repository::NgWordRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("name", user.name.as_str().into()),
            ("display_name", user.display_name.as_str().into()),
            ("password", user.password.as_str().into()),
            ("description", user.description.as_str().into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("user_id", 1i64.into()),
            ("title", stream.title.as_str().into()),
            ("description", stream.description.as_str().into()),
            ("playlist_url", stream.playlist_url.as_str().into()),
            ("thumbnail_url", stream.thumbnail_url.as_str().into()),
            ("start_at", stream.start_at.into()),
            ("end_at", stream.end_at.into()),
        ]);
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

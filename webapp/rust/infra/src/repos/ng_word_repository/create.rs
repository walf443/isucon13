use crate::qbey_support::bind_qbey_values;
use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::ng_word::{CreateNgWord, NgWord};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("name", "test".into()), ("display_name", "Test".into()), ("password", "pw".into()), ("description", "desc".into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("title", "t1".into()), ("description", "d1".into()), ("playlist_url", "http://p".into()), ("thumbnail_url", "http://t".into()), ("start_at", 1000i64.into()), ("end_at", 2000i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let input = CreateNgWord {
        user_id: UserId::new(1),
        livestream_id: LivestreamId::new(1),
        word: "badword".to_string(),
        created_at: 300,
    };

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

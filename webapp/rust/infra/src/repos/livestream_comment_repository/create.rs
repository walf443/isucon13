use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamComment};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
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

    let input = CreateLivestreamComment {
        user_id: UserId::new(1),
        livestream_id: LivestreamId::new(1),
        comment: "hello world".to_string(),
        tip: 500,
        created_at: 1234,
    };

    let repo = LivestreamCommentRepositoryInfra {};
    let comment_id = repo.create(&mut tx, &input).await.unwrap();

    let got: LivestreamComment = sqlx::query_as("SELECT * FROM livecomments WHERE id = ?")
        .bind(&comment_id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();

    assert_eq!(got.id, comment_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.comment, input.comment);
    assert_eq!(got.tip, input.tip);
    assert_eq!(got.created_at, input.created_at);
}

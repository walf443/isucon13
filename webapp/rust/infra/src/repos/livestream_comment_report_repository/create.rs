use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::user::TABLE_USERS;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::LivestreamCommentId;
use isupipe_core::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReport,
};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
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

    {
        let mut ins = qbey(TABLE_LIVECOMMENTS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("comment", "hello".into()), ("tip", 0i64.into()), ("created_at", 100i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let input = CreateLivestreamCommentReport {
        user_id: UserId::new(1),
        livestream_id: LivestreamId::new(1),
        livestream_comment_id: LivestreamCommentId::new(1),
        created_at: 200,
    };

    let repo = LivestreamCommentReportRepositoryInfra {};
    let report_id = repo.create(&mut tx, &input).await.unwrap();

    let got: LivestreamCommentReport =
        sqlx::query_as("SELECT * FROM livecomment_reports WHERE id = ?")
            .bind(&report_id)
            .fetch_one(&mut *tx)
            .await
            .unwrap();

    assert_eq!(got.id, report_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.livestream_comment_id, input.livestream_comment_id);
    assert_eq!(got.created_at, input.created_at);
}

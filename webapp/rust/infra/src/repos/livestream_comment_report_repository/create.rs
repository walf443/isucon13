use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::LivestreamCommentId;
use isupipe_core::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReport,
};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'test', 'Test', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't1', 'd1', 'http://p', 'http://t', 1000, 2000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livecomments (id, user_id, livestream_id, comment, tip, created_at) VALUES (1, 1, 1, 'hello', 0, 100)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

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

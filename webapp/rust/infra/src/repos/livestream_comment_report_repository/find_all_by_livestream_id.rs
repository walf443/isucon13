use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;

#[tokio::test]
async fn empty_case() {
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

    let repo = LivestreamCommentReportRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_livestream_id() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'test', 'Test', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't1', 'd1', 'http://p', 'http://t', 1000, 2000), (2, 1, 't2', 'd2', 'http://p2', 'http://t2', 3000, 4000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livecomments (id, user_id, livestream_id, comment, tip, created_at) VALUES (1, 1, 1, 'a', 0, 100)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livecomment_reports (id, user_id, livestream_id, livecomment_id, created_at) VALUES (1, 1, 1, 1, 100), (2, 1, 1, 1, 200), (3, 1, 2, 1, 300)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let repo = LivestreamCommentReportRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 2);

    let result = repo
        .find_all_by_livestream_id(&mut tx, &LivestreamId::new(2))
        .await
        .unwrap();
    assert_eq!(result.len(), 1);
}

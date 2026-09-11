use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::livestream_comment_report::TABLE_LIVECOMMENT_REPORTS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{InsertCommentSetup, InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamCommentId};
use isupipe_core::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReport,
};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let comment: CreateLivestreamComment = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVECOMMENTS.table()).into_insert();
        ins.add_value(&InsertCommentSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            comment: &comment,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let mut input: CreateLivestreamCommentReport = Faker.fake();
    input.user_id = UserId::new(1);
    input.livestream_id = LivestreamId::new(1);
    input.livestream_comment_id = LivestreamCommentId::new(1);

    let repo = LivestreamCommentReportRepositoryInfra {};
    let report_id = repo.create(&mut tx, &input).await.unwrap();

    let t = &TABLE_LIVECOMMENT_REPORTS;
    let mut q = qbey(t.table());
    q.and_where(t.id().eq(*report_id.inner()));
    let (sql, binds) = q.into_sql();
    let got: LivestreamCommentReport = bind_qbey_values!(
        sqlx::query_as::<_, LivestreamCommentReport>(sqlx::AssertSqlSafe(sql)),
        binds
    )
    .fetch_one(&mut *tx)
    .await
    .unwrap();

    assert_eq!(got.id, report_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.livestream_comment_id, input.livestream_comment_id);
    assert_eq!(got.created_at, input.created_at);
}

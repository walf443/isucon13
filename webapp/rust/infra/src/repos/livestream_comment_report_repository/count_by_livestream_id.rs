use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::livestream_comment_report::TABLE_LIVECOMMENT_REPORTS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{
    InsertCommentSetup, InsertLivestreamSetup, InsertReportSetup, InsertUserSetup,
};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_comment::CreateLivestreamComment;
use isupipe_core::models::livestream_comment_report::CreateLivestreamCommentReport;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn zero_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
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
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamCommentReportRepositoryInfra {};
    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 0);
}

#[tokio::test]
async fn counts_correctly() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream1,
        });
        ins.add_value(&InsertLivestreamSetup {
            id: 2,
            user_id: 1,
            stream: &stream2,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let comment1: CreateLivestreamComment = Faker.fake();
    let comment2: CreateLivestreamComment = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVECOMMENTS.table()).into_insert();
        ins.add_value(&InsertCommentSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            comment: &comment1,
        });
        ins.add_value(&InsertCommentSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            comment: &comment2,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVECOMMENT_REPORTS.table()).into_insert();
        let r1: CreateLivestreamCommentReport = Faker.fake();
        ins.add_value(&InsertReportSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            livecomment_id: 1,
            report: &r1,
        });
        let r2: CreateLivestreamCommentReport = Faker.fake();
        ins.add_value(&InsertReportSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            livecomment_id: 2,
            report: &r2,
        });
        let r3: CreateLivestreamCommentReport = Faker.fake();
        ins.add_value(&InsertReportSetup {
            id: 3,
            user_id: 1,
            livestream_id: 2,
            livecomment_id: 1,
            report: &r3,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamCommentReportRepositoryInfra {};
    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 2);

    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(2))
        .await
        .unwrap();
    assert_eq!(result, 1);
}

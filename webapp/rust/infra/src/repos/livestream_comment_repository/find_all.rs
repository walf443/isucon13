use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertCommentSetup, InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::livestream_comment::CreateLivestreamComment;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn not_empty_case() {
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

    let c1: CreateLivestreamComment = Faker.fake();
    let c2: CreateLivestreamComment = Faker.fake();
    {
        InsertCommentSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            comment: &c1,
        }
        .insert(&mut tx)
        .await;
        InsertCommentSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            comment: &c2,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(result.len(), 2);
}

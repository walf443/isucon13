use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertCommentSetup, InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_comment::CreateLivestreamComment;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn returns_ordered_by_created_at_desc() {
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

    {
        let c1: CreateLivestreamComment = CreateLivestreamComment {
            created_at: 100,
            ..Faker.fake()
        };
        InsertCommentSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            comment: &c1,
        }
        .insert(&mut tx)
        .await;
        let c2: CreateLivestreamComment = CreateLivestreamComment {
            created_at: 300,
            ..Faker.fake()
        };
        InsertCommentSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            comment: &c2,
        }
        .insert(&mut tx)
        .await;
        let c3: CreateLivestreamComment = CreateLivestreamComment {
            created_at: 200,
            ..Faker.fake()
        };
        InsertCommentSetup {
            id: 3,
            user_id: 1,
            livestream_id: 1,
            comment: &c3,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_order_by_created_at(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 3);
    assert_eq!(result[0].created_at, 300);
    assert_eq!(result[1].created_at, 200);
    assert_eq!(result[2].created_at, 100);
}

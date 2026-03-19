use crate::qbey_support::bind_qbey_values;
use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::reaction::TABLE_REACTIONS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{InsertLivestreamSetup, InsertReactionSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::reaction::CreateReaction;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::reaction_repository::ReactionRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn returns_limited_rows() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.to_sql();
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
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    {
        let mut ins = qbey(TABLE_REACTIONS.table()).into_insert();
        let r1: CreateReaction = CreateReaction {
            created_at: 100,
            ..Faker.fake()
        };
        ins.add_value(&InsertReactionSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            reaction: &r1,
        });
        let r2: CreateReaction = CreateReaction {
            created_at: 300,
            ..Faker.fake()
        };
        ins.add_value(&InsertReactionSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            reaction: &r2,
        });
        let r3: CreateReaction = CreateReaction {
            created_at: 200,
            ..Faker.fake()
        };
        ins.add_value(&InsertReactionSetup {
            id: 3,
            user_id: 1,
            livestream_id: 1,
            reaction: &r3,
        });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_limit(&mut tx, &LivestreamId::new(1), 2)
        .await
        .unwrap();
    assert_eq!(result.len(), 2);
    assert_eq!(result[0].created_at, 300);
    assert_eq!(result[1].created_at, 200);
}

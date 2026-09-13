//! リポジトリのテストで使う、テスト DB の組み立てとセットアップ用 INSERT のヘルパ。
//!
//! 各 `Insert*Setup` は既存のテストコードとの互換のために struct literal で組み立て、
//! `.insert(&mut tx).await` で toasty のモデル経由で INSERT する。

use crate::tables::icon::IconRow;
use crate::tables::livestream::LivestreamRow;
use crate::tables::livestream_comment::LivestreamCommentRow;
use crate::tables::livestream_comment_report::LivestreamCommentReportRow;
use crate::tables::livestream_tag::LivestreamTagRow;
use crate::tables::livestream_viewers_history::LivestreamViewersHistoryRow;
use crate::tables::ng_word::NgWordRow;
use crate::tables::reaction::ReactionRow;
use crate::tables::reservation_slot::ReservationSlotRow;
use crate::tables::tag::TagRow;
use crate::tables::theme::ThemeRow;
use crate::tables::user::UserRow;
use isupipe_core::db::{DBConn, DBPool};
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::livestream_comment::CreateLivestreamComment;
use isupipe_core::models::livestream_comment_report::CreateLivestreamCommentReport;
use isupipe_core::models::ng_word::CreateNgWord;
use isupipe_core::models::reaction::CreateReaction;
use isupipe_core::models::reservation_slot::ReservationSlot;
use isupipe_core::models::tag::Tag;
use isupipe_core::models::theme::Theme;

/// モデル登録済みのテスト用 `Db` を返す。テストごとに独立したプールを作る。
pub async fn get_db_pool() -> DBPool {
    let url = isupipe_core::db::get_test_db_url().await;
    crate::db::build_db(&url, 2).await.unwrap()
}

pub struct InsertUserSetup<'a> {
    pub id: i64,
    pub user: &'a isupipe_core::models::user::CreateUser,
}

impl InsertUserSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> UserRow {
        UserRow::create()
            .id(self.id)
            .name(&self.user.name)
            .display_name(&self.user.display_name)
            .hashed_password(&self.user.password)
            .description(&self.user.description)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertLivestreamSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub stream: &'a CreateLivestream,
}

impl InsertLivestreamSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> LivestreamRow {
        LivestreamRow::create()
            .id(self.id)
            .user_id(self.user_id)
            .title(&self.stream.title)
            .description(&self.stream.description)
            .playlist_url(&self.stream.playlist_url)
            .thumbnail_url(&self.stream.thumbnail_url)
            .start_at(self.stream.start_at)
            .end_at(self.stream.end_at)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertReactionSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub reaction: &'a CreateReaction,
}

impl InsertReactionSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> ReactionRow {
        ReactionRow::create()
            .id(self.id)
            .user_id(self.user_id)
            .livestream_id(self.livestream_id)
            .emoji_name(&self.reaction.emoji_name)
            .created_at(self.reaction.created_at)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertCommentSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub comment: &'a CreateLivestreamComment,
}

impl InsertCommentSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> LivestreamCommentRow {
        LivestreamCommentRow::create()
            .id(self.id)
            .user_id(self.user_id)
            .livestream_id(self.livestream_id)
            .comment(&self.comment.comment)
            .tip(self.comment.tip)
            .created_at(self.comment.created_at)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertReportSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub livecomment_id: i64,
    pub report: &'a CreateLivestreamCommentReport,
}

impl InsertReportSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> LivestreamCommentReportRow {
        LivestreamCommentReportRow::create()
            .id(self.id)
            .user_id(self.user_id)
            .livestream_id(self.livestream_id)
            .livecomment_id(self.livecomment_id)
            .created_at(self.report.created_at)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertNgWordSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub ng_word: &'a CreateNgWord,
}

impl InsertNgWordSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> NgWordRow {
        NgWordRow::create()
            .id(self.id)
            .user_id(self.user_id)
            .livestream_id(self.livestream_id)
            .word(&self.ng_word.word)
            .created_at(self.ng_word.created_at)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertTagSetup<'a> {
    pub id: i64,
    pub tag: &'a Tag,
}

impl InsertTagSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> TagRow {
        TagRow::create()
            .id(self.id)
            .name(self.tag.name.inner())
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertLivestreamTagSetup {
    pub id: i64,
    pub livestream_id: i64,
    pub tag_id: i64,
}

impl InsertLivestreamTagSetup {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> LivestreamTagRow {
        LivestreamTagRow::create()
            .id(self.id)
            .livestream_id(self.livestream_id)
            .tag_id(self.tag_id)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertViewersHistorySetup {
    pub user_id: i64,
    pub livestream_id: i64,
    pub created_at: i64,
}

impl InsertViewersHistorySetup {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> LivestreamViewersHistoryRow {
        LivestreamViewersHistoryRow::create()
            .user_id(self.user_id)
            .livestream_id(self.livestream_id)
            .created_at(self.created_at)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertReservationSlotSetup<'a> {
    pub slot: &'a ReservationSlot,
}

impl InsertReservationSlotSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> ReservationSlotRow {
        ReservationSlotRow::create()
            .id(self.slot.id.inner())
            .slot(self.slot.slot)
            .start_at(self.slot.start_at)
            .end_at(self.slot.end_at)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertThemeSetup<'a> {
    pub theme: &'a Theme,
}

impl InsertThemeSetup<'_> {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> ThemeRow {
        ThemeRow::create()
            .id(self.theme.id.inner())
            .user_id(self.theme.user_id.inner())
            .dark_mode(self.theme.dark_mode)
            .exec(conn)
            .await
            .unwrap()
    }
}

pub struct InsertIconSetup {
    pub id: i64,
    pub user_id: i64,
    pub image: Vec<u8>,
}

impl InsertIconSetup {
    pub async fn insert(&self, conn: &mut DBConn<'_>) -> IconRow {
        IconRow::create()
            .id(self.id)
            .user_id(self.user_id)
            .image(self.image.clone())
            .exec(conn)
            .await
            .unwrap()
    }
}

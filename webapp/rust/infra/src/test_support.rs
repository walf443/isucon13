//! リポジトリのテストで使う、セットアップ用 INSERT の行定義。
//!
//! 各 `Insert*Setup` は `ToInsertRow` を実装し、`ins.add_value(&InsertXSetup { .. })` で使う。
//! カラムはテーブル定義 (`crate::tables`) の型付きカラム経由で指定するので、
//! カラム名の typo や型の取り違えはコンパイルエラーになる。

use crate::tables::icon::TABLE_ICONS;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::livestream_comment_report::TABLE_LIVECOMMENT_REPORTS;
use crate::tables::livestream_tag::TABLE_LIVESTREAM_TAGS;
use crate::tables::livestream_viewers_history::TABLE_LIVESTREAM_VIEWERS_HISTORY;
use crate::tables::ng_word::TABLE_NG_WORDS;
use crate::tables::reaction::TABLE_REACTIONS;
use crate::tables::reservation_slot::TABLE_RESERVATION_SLOTS;
use crate::tables::tag::TABLE_TAGS;
use crate::tables::theme::TABLE_THEMES;
use crate::tables::user::TABLE_USERS;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamCommentId};
use isupipe_core::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReportId,
};
use isupipe_core::models::livestream_tag::LivestreamTagId;
use isupipe_core::models::ng_word::{CreateNgWord, NgWordId};
use isupipe_core::models::reaction::{CreateReaction, ReactionId};
use isupipe_core::models::reservation_slot::ReservationSlot;
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::models::theme::Theme;
use isupipe_core::models::user::{UserId, UserName};
use qbey::Value;

pub struct InsertUserSetup<'a> {
    pub id: i64,
    pub user: &'a isupipe_core::models::user::CreateUser,
}

impl qbey::ToInsertRow<Value, String> for InsertUserSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_USERS;
        vec![
            t.id().value(UserId::new(self.id)),
            t.name().value(UserName::new(self.user.name.clone())),
            t.display_name().value(&self.user.display_name),
            t.password().value(&self.user.password),
            t.description().value(&self.user.description),
        ]
    }
}

pub struct InsertLivestreamSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub stream: &'a CreateLivestream,
}

impl qbey::ToInsertRow<Value, String> for InsertLivestreamSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_LIVESTREAMS;
        vec![
            t.id().value(LivestreamId::new(self.id)),
            t.user_id().value(UserId::new(self.user_id)),
            t.title().value(&self.stream.title),
            t.description().value(&self.stream.description),
            t.playlist_url().value(&self.stream.playlist_url),
            t.thumbnail_url().value(&self.stream.thumbnail_url),
            t.start_at().value(self.stream.start_at),
            t.end_at().value(self.stream.end_at),
        ]
    }
}

pub struct InsertReactionSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub reaction: &'a CreateReaction,
}

impl qbey::ToInsertRow<Value, String> for InsertReactionSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_REACTIONS;
        vec![
            t.id().value(ReactionId::new(self.id)),
            t.user_id().value(UserId::new(self.user_id)),
            t.livestream_id()
                .value(LivestreamId::new(self.livestream_id)),
            t.emoji_name().value(&self.reaction.emoji_name),
            t.created_at().value(self.reaction.created_at),
        ]
    }
}

pub struct InsertCommentSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub comment: &'a CreateLivestreamComment,
}

impl qbey::ToInsertRow<Value, String> for InsertCommentSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_LIVECOMMENTS;
        vec![
            t.id().value(LivestreamCommentId::new(self.id)),
            t.user_id().value(UserId::new(self.user_id)),
            t.livestream_id()
                .value(LivestreamId::new(self.livestream_id)),
            t.comment().value(&self.comment.comment),
            t.tip().value(self.comment.tip),
            t.created_at().value(self.comment.created_at),
        ]
    }
}

pub struct InsertReportSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub livecomment_id: i64,
    pub report: &'a CreateLivestreamCommentReport,
}

impl qbey::ToInsertRow<Value, String> for InsertReportSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_LIVECOMMENT_REPORTS;
        vec![
            t.id().value(LivestreamCommentReportId::new(self.id)),
            t.user_id().value(UserId::new(self.user_id)),
            t.livestream_id()
                .value(LivestreamId::new(self.livestream_id)),
            t.livecomment_id()
                .value(LivestreamCommentId::new(self.livecomment_id)),
            t.created_at().value(self.report.created_at),
        ]
    }
}

pub struct InsertNgWordSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub ng_word: &'a CreateNgWord,
}

impl qbey::ToInsertRow<Value, String> for InsertNgWordSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_NG_WORDS;
        vec![
            t.id().value(NgWordId::new(self.id)),
            t.user_id().value(UserId::new(self.user_id)),
            t.livestream_id()
                .value(LivestreamId::new(self.livestream_id)),
            t.word().value(&self.ng_word.word),
            t.created_at().value(self.ng_word.created_at),
        ]
    }
}

pub struct InsertTagSetup<'a> {
    pub id: i64,
    pub tag: &'a Tag,
}

impl qbey::ToInsertRow<Value, String> for InsertTagSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_TAGS;
        vec![
            t.id().value(TagId::new(self.id)),
            t.name().value(&self.tag.name),
        ]
    }
}

pub struct InsertLivestreamTagSetup {
    pub id: i64,
    pub livestream_id: i64,
    pub tag_id: i64,
}

impl qbey::ToInsertRow<Value, String> for InsertLivestreamTagSetup {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_LIVESTREAM_TAGS;
        vec![
            t.id().value(LivestreamTagId::new(self.id)),
            t.livestream_id()
                .value(LivestreamId::new(self.livestream_id)),
            t.tag_id().value(TagId::new(self.tag_id)),
        ]
    }
}

pub struct InsertViewersHistorySetup {
    pub user_id: i64,
    pub livestream_id: i64,
    pub created_at: i64,
}

impl qbey::ToInsertRow<Value, String> for InsertViewersHistorySetup {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_LIVESTREAM_VIEWERS_HISTORY;
        vec![
            t.user_id().value(UserId::new(self.user_id)),
            t.livestream_id()
                .value(LivestreamId::new(self.livestream_id)),
            t.created_at().value(self.created_at),
        ]
    }
}

pub struct InsertReservationSlotSetup<'a> {
    pub slot: &'a ReservationSlot,
}

impl qbey::ToInsertRow<Value, String> for InsertReservationSlotSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_RESERVATION_SLOTS;
        vec![
            t.id().value(&self.slot.id),
            t.slot().value(self.slot.slot),
            t.start_at().value(self.slot.start_at),
            t.end_at().value(self.slot.end_at),
        ]
    }
}

pub struct InsertThemeSetup<'a> {
    pub theme: &'a Theme,
}

impl qbey::ToInsertRow<Value, String> for InsertThemeSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_THEMES;
        vec![
            t.id().value(&self.theme.id),
            t.user_id().value(&self.theme.user_id),
            t.dark_mode().value(self.theme.dark_mode),
        ]
    }
}

pub struct InsertIconSetup {
    pub id: i64,
    pub user_id: i64,
    pub image: Vec<u8>,
}

impl qbey::ToInsertRow<Value, String> for InsertIconSetup {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let t = &TABLE_ICONS;
        vec![
            t.id().value(self.id),
            t.user_id().value(UserId::new(self.user_id)),
            t.image().value(&self.image),
        ]
    }
}

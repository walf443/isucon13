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
        let mut row = TABLE_USERS.row();
        row.id(UserId::new(self.id))
            .name(UserName::new(self.user.name.clone()))
            .display_name(&self.user.display_name)
            .password(&self.user.password)
            .description(&self.user.description);
        row.to_insert_row()
    }
}

pub struct InsertLivestreamSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub stream: &'a CreateLivestream,
}

impl qbey::ToInsertRow<Value, String> for InsertLivestreamSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let mut row = TABLE_LIVESTREAMS.row();
        row.id(LivestreamId::new(self.id))
            .user_id(UserId::new(self.user_id))
            .title(&self.stream.title)
            .description(&self.stream.description)
            .playlist_url(&self.stream.playlist_url)
            .thumbnail_url(&self.stream.thumbnail_url)
            .start_at(self.stream.start_at)
            .end_at(self.stream.end_at);
        row.to_insert_row()
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
        let mut row = TABLE_REACTIONS.row();
        row.id(ReactionId::new(self.id))
            .user_id(UserId::new(self.user_id))
            .livestream_id(LivestreamId::new(self.livestream_id))
            .emoji_name(&self.reaction.emoji_name)
            .created_at(self.reaction.created_at);
        row.to_insert_row()
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
        let mut row = TABLE_LIVECOMMENTS.row();
        row.id(LivestreamCommentId::new(self.id))
            .user_id(UserId::new(self.user_id))
            .livestream_id(LivestreamId::new(self.livestream_id))
            .comment(&self.comment.comment)
            .tip(self.comment.tip)
            .created_at(self.comment.created_at);
        row.to_insert_row()
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
        let mut row = TABLE_LIVECOMMENT_REPORTS.row();
        row.id(LivestreamCommentReportId::new(self.id))
            .user_id(UserId::new(self.user_id))
            .livestream_id(LivestreamId::new(self.livestream_id))
            .livecomment_id(LivestreamCommentId::new(self.livecomment_id))
            .created_at(self.report.created_at);
        row.to_insert_row()
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
        let mut row = TABLE_NG_WORDS.row();
        row.id(NgWordId::new(self.id))
            .user_id(UserId::new(self.user_id))
            .livestream_id(LivestreamId::new(self.livestream_id))
            .word(&self.ng_word.word)
            .created_at(self.ng_word.created_at);
        row.to_insert_row()
    }
}

pub struct InsertTagSetup<'a> {
    pub id: i64,
    pub tag: &'a Tag,
}

impl qbey::ToInsertRow<Value, String> for InsertTagSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let mut row = TABLE_TAGS.row();
        row.id(TagId::new(self.id)).name(&self.tag.name);
        row.to_insert_row()
    }
}

pub struct InsertLivestreamTagSetup {
    pub id: i64,
    pub livestream_id: i64,
    pub tag_id: i64,
}

impl qbey::ToInsertRow<Value, String> for InsertLivestreamTagSetup {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let mut row = TABLE_LIVESTREAM_TAGS.row();
        row.id(LivestreamTagId::new(self.id))
            .livestream_id(LivestreamId::new(self.livestream_id))
            .tag_id(TagId::new(self.tag_id));
        row.to_insert_row()
    }
}

pub struct InsertViewersHistorySetup {
    pub user_id: i64,
    pub livestream_id: i64,
    pub created_at: i64,
}

impl qbey::ToInsertRow<Value, String> for InsertViewersHistorySetup {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let mut row = TABLE_LIVESTREAM_VIEWERS_HISTORY.row();
        row.user_id(UserId::new(self.user_id))
            .livestream_id(LivestreamId::new(self.livestream_id))
            .created_at(self.created_at);
        row.to_insert_row()
    }
}

pub struct InsertReservationSlotSetup<'a> {
    pub slot: &'a ReservationSlot,
}

impl qbey::ToInsertRow<Value, String> for InsertReservationSlotSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let mut row = TABLE_RESERVATION_SLOTS.row();
        row.id(&self.slot.id)
            .slot(self.slot.slot)
            .start_at(self.slot.start_at)
            .end_at(self.slot.end_at);
        row.to_insert_row()
    }
}

pub struct InsertThemeSetup<'a> {
    pub theme: &'a Theme,
}

impl qbey::ToInsertRow<Value, String> for InsertThemeSetup<'_> {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let mut row = TABLE_THEMES.row();
        row.id(&self.theme.id)
            .user_id(&self.theme.user_id)
            .dark_mode(self.theme.dark_mode);
        row.to_insert_row()
    }
}

pub struct InsertIconSetup {
    pub id: i64,
    pub user_id: i64,
    pub image: Vec<u8>,
}

impl qbey::ToInsertRow<Value, String> for InsertIconSetup {
    fn to_insert_row(&self) -> Vec<(String, Value)> {
        let mut row = TABLE_ICONS.row();
        row.id(self.id)
            .user_id(UserId::new(self.user_id))
            .image(&self.image);
        row.to_insert_row()
    }
}

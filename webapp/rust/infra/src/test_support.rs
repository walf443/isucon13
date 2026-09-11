use isupipe_core::models::icon::CreateIcon;
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::livestream_comment::CreateLivestreamComment;
use isupipe_core::models::livestream_comment_report::CreateLivestreamCommentReport;
use isupipe_core::models::ng_word::CreateNgWord;
use isupipe_core::models::reaction::CreateReaction;
use isupipe_core::models::reservation_slot::ReservationSlot;
use isupipe_core::models::tag::Tag;
use isupipe_core::models::theme::Theme;

pub struct InsertUserSetup<'a> {
    pub id: i64,
    pub user: &'a isupipe_core::models::user::CreateUser,
}

impl qbey::ToInsertRow<qbey::Value> for InsertUserSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", self.id.into()),
            ("name", self.user.name.as_str().into()),
            ("display_name", self.user.display_name.as_str().into()),
            ("password", self.user.password.as_str().into()),
            ("description", self.user.description.as_str().into()),
        ]
    }
}

pub struct InsertLivestreamSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub stream: &'a CreateLivestream,
}

impl qbey::ToInsertRow<qbey::Value> for InsertLivestreamSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", self.id.into()),
            ("user_id", self.user_id.into()),
            ("title", self.stream.title.as_str().into()),
            ("description", self.stream.description.as_str().into()),
            ("playlist_url", self.stream.playlist_url.as_str().into()),
            ("thumbnail_url", self.stream.thumbnail_url.as_str().into()),
            ("start_at", self.stream.start_at.into()),
            ("end_at", self.stream.end_at.into()),
        ]
    }
}

pub struct InsertReactionSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub reaction: &'a CreateReaction,
}

impl qbey::ToInsertRow<qbey::Value> for InsertReactionSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", self.id.into()),
            ("user_id", self.user_id.into()),
            ("livestream_id", self.livestream_id.into()),
            ("emoji_name", self.reaction.emoji_name.as_str().into()),
            ("created_at", self.reaction.created_at.into()),
        ]
    }
}

pub struct InsertCommentSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub comment: &'a CreateLivestreamComment,
}

impl qbey::ToInsertRow<qbey::Value> for InsertCommentSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", self.id.into()),
            ("user_id", self.user_id.into()),
            ("livestream_id", self.livestream_id.into()),
            ("comment", self.comment.comment.as_str().into()),
            ("tip", self.comment.tip.into()),
            ("created_at", self.comment.created_at.into()),
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

impl qbey::ToInsertRow<qbey::Value> for InsertReportSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", self.id.into()),
            ("user_id", self.user_id.into()),
            ("livestream_id", self.livestream_id.into()),
            ("livecomment_id", self.livecomment_id.into()),
            ("created_at", self.report.created_at.into()),
        ]
    }
}

pub struct InsertNgWordSetup<'a> {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub ng_word: &'a CreateNgWord,
}

impl qbey::ToInsertRow<qbey::Value> for InsertNgWordSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", self.id.into()),
            ("user_id", self.user_id.into()),
            ("livestream_id", self.livestream_id.into()),
            ("word", self.ng_word.word.as_str().into()),
            ("created_at", self.ng_word.created_at.into()),
        ]
    }
}

pub struct InsertTagSetup<'a> {
    pub id: i64,
    pub tag: &'a Tag,
}

impl qbey::ToInsertRow<qbey::Value> for InsertTagSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", self.id.into()),
            ("name", self.tag.name.inner().as_str().into()),
        ]
    }
}

pub struct InsertLivestreamTagSetup {
    pub id: i64,
    pub livestream_id: i64,
    pub tag_id: i64,
}

impl qbey::ToInsertRow<qbey::Value> for InsertLivestreamTagSetup {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", self.id.into()),
            ("livestream_id", self.livestream_id.into()),
            ("tag_id", self.tag_id.into()),
        ]
    }
}

pub struct InsertViewersHistorySetup {
    pub user_id: i64,
    pub livestream_id: i64,
    pub created_at: i64,
}

impl qbey::ToInsertRow<qbey::Value> for InsertViewersHistorySetup {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("user_id", self.user_id.into()),
            ("livestream_id", self.livestream_id.into()),
            ("created_at", self.created_at.into()),
        ]
    }
}

pub struct InsertReservationSlotSetup<'a> {
    pub slot: &'a ReservationSlot,
}

impl qbey::ToInsertRow<qbey::Value> for InsertReservationSlotSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", (*self.slot.id.inner()).into()),
            ("slot", self.slot.slot.into()),
            ("start_at", self.slot.start_at.into()),
            ("end_at", self.slot.end_at.into()),
        ]
    }
}

pub struct InsertThemeSetup<'a> {
    pub theme: &'a Theme,
}

impl qbey::ToInsertRow<qbey::Value> for InsertThemeSetup<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("id", (*self.theme.id.inner()).into()),
            ("user_id", (*self.theme.user_id.inner()).into()),
            ("dark_mode", self.theme.dark_mode.into()),
        ]
    }
}

pub struct InsertIconSetup {
    pub id: i64,
    pub user_id: i64,
    pub image: Vec<u8>,
}

impl qbey::ToInsertRow<crate::qbey_support::SQLValue> for InsertIconSetup {
    fn to_insert_row(&self) -> Vec<(&'static str, crate::qbey_support::SQLValue)> {
        vec![
            ("id", self.id.into()),
            ("user_id", self.user_id.into()),
            ("image", self.image.clone().into()),
        ]
    }
}

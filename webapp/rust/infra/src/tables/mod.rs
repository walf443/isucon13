//! toasty のモデル定義 (DB のテーブルと 1:1 で対応する行構造体)。
//!
//! core のドメインモデルは kubetsu の `Id<T, U>` を使っているため、
//! ここでは素の `i64` / `String` で行を受け取り、`From` でドメインモデルへ変換する。

pub mod icon;
pub mod livestream;
pub mod livestream_comment;
pub mod livestream_comment_report;
pub mod livestream_tag;
pub mod livestream_viewers_history;
pub mod ng_word;
pub mod reaction;
pub mod reservation_slot;
pub mod tag;
pub mod theme;
pub mod user;

/// この crate で定義している全モデルを toasty に登録するための `ModelSet`。
pub fn models() -> toasty::ModelSet {
    toasty::models!(
        icon::IconRow,
        livestream::LivestreamRow,
        livestream_comment::LivestreamCommentRow,
        livestream_comment_report::LivestreamCommentReportRow,
        livestream_tag::LivestreamTagRow,
        livestream_viewers_history::LivestreamViewersHistoryRow,
        ng_word::NgWordRow,
        reaction::ReactionRow,
        reservation_slot::ReservationSlotRow,
        tag::TagRow,
        theme::ThemeRow,
        user::UserRow,
    )
}

macro_rules! sqipe_schema {
    ($struct_name:ident, $table_name:expr, [$($col:ident),* $(,)?]) => {
        use sqipe::{table, Col, TableRef};

        pub struct $struct_name {
            alias: Option<&'static str>,
        }

        impl $struct_name {
            pub const fn new() -> Self {
                $struct_name { alias: None }
            }

            pub fn table_name(&self) -> &'static str {
                $table_name
            }

            pub fn table(&self) -> TableRef {
                table(self.alias.unwrap_or(self.table_name()))
            }

            pub fn as_(&self, alias: &'static str) -> Self {
                $struct_name { alias: Some(alias) }
            }

            $(
                pub fn $col(&self) -> Col {
                    self.table().col(stringify!($col))
                }
            )*

            pub fn all_columns(&self) -> Vec<Col> {
                vec![$(self.$col()),*]
            }
        }
    };
}

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

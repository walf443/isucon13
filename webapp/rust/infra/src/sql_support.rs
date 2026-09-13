//! `toasty::sql::query` (raw SQL) の結果から値を取り出すヘルパ。
//!
//! 集計 (SUM / MAX / COUNT with JOIN) や `FOR UPDATE` など toasty のクエリビルダで
//! 表現できないクエリは raw SQL で実行し、ここで `Value` から取り出す。

use toasty::stmt::Value;

/// 1 行目の先頭カラムを取り出す。
fn first_column(rows: Vec<Value>) -> toasty::Result<Option<Value>> {
    let Some(row) = rows.into_iter().next() else {
        return Ok(None);
    };
    match row {
        Value::Record(record) => Ok(record.fields.into_iter().next()),
        other => Err(toasty::Error::invalid_result(format!(
            "expected a record row, got {other:?}"
        ))),
    }
}

/// 1 行目の先頭カラムを `i64` として取り出す。行が無い場合はエラー。
///
/// MySQL の SUM() は DECIMAL になるので、呼び出し側で `CAST(... AS SIGNED)` すること。
pub(crate) fn scalar_i64(rows: Vec<Value>) -> toasty::Result<i64> {
    match first_column(rows)? {
        Some(Value::I64(n)) => Ok(n),
        Some(Value::U64(n)) => i64::try_from(n)
            .map_err(|_| toasty::Error::invalid_result(format!("value out of range: {n}"))),
        Some(Value::I32(n)) => Ok(n.into()),
        Some(Value::U32(n)) => Ok(n.into()),
        Some(Value::Null) => Ok(0),
        Some(other) => Err(toasty::Error::invalid_result(format!(
            "expected an integer, got {other:?}"
        ))),
        None => Err(toasty::Error::invalid_result("expected one row, got none")),
    }
}

/// 1 行目の先頭カラムを `String` として取り出す。行が無い場合は `None`。
pub(crate) fn optional_scalar_string(rows: Vec<Value>) -> toasty::Result<Option<String>> {
    match first_column(rows)? {
        Some(Value::String(s)) => Ok(Some(s)),
        Some(Value::Null) | None => Ok(None),
        Some(other) => Err(toasty::Error::invalid_result(format!(
            "expected a string, got {other:?}"
        ))),
    }
}

/// Bind `qbey::Value` parameters to a sqlx query.
///
/// Works with `sqlx::query_as`, `sqlx::query_scalar`, etc.
/// The query type must have a `bind` method that accepts `String`, `i64`, `f64`, and `bool`.
macro_rules! bind_qbey_values {
    ($query:expr, $binds:expr) => {{
        let mut q = $query;
        for v in $binds {
            q = match v {
                qbey::Value::String(s) => q.bind(s),
                qbey::Value::Int(n) => q.bind(n),
                qbey::Value::Float(f) => q.bind(f),
                qbey::Value::Bool(b) => q.bind(b),
                qbey::Value::Bytes(b) => q.bind(b),
            };
        }
        q
    }};
}

pub(crate) use bind_qbey_values;

/// Project-wide custom value type for qbey queries that need types beyond `qbey::Value`.
///
/// Use with `qbey_mysql::qbey_with::<SQLValue>(table)` when a query involves
/// types not supported by `qbey::Value` (e.g., `Vec<u8>`).
#[derive(Debug, Clone)]
pub(crate) enum SQLValue {
    String(String),
    Int(i64),
    Float(f64),
    Bool(bool),
    Bytes(Vec<u8>),
}

impl From<&str> for SQLValue {
    fn from(s: &str) -> Self {
        SQLValue::String(s.to_string())
    }
}

impl From<String> for SQLValue {
    fn from(s: String) -> Self {
        SQLValue::String(s)
    }
}

impl From<i64> for SQLValue {
    fn from(n: i64) -> Self {
        SQLValue::Int(n)
    }
}

impl From<i32> for SQLValue {
    fn from(n: i32) -> Self {
        SQLValue::Int(n as i64)
    }
}

impl From<f64> for SQLValue {
    fn from(n: f64) -> Self {
        SQLValue::Float(n)
    }
}

impl From<bool> for SQLValue {
    fn from(b: bool) -> Self {
        SQLValue::Bool(b)
    }
}

impl From<Vec<u8>> for SQLValue {
    fn from(b: Vec<u8>) -> Self {
        SQLValue::Bytes(b)
    }
}

/// Bind `SQLValue` parameters to a sqlx query.
macro_rules! bind_sql_values {
    ($query:expr, $binds:expr) => {{
        let mut q = $query;
        for v in $binds {
            q = match v {
                $crate::qbey_support::SQLValue::String(s) => q.bind(s),
                $crate::qbey_support::SQLValue::Int(n) => q.bind(n),
                $crate::qbey_support::SQLValue::Float(f) => q.bind(f),
                $crate::qbey_support::SQLValue::Bool(b) => q.bind(b),
                $crate::qbey_support::SQLValue::Bytes(b) => q.bind(b),
            };
        }
        q
    }};
}

pub(crate) use bind_sql_values;

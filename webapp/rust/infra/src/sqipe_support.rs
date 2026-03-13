/// Bind `sqipe::Value` parameters to a sqlx query.
///
/// Works with `sqlx::query_as`, `sqlx::query_scalar`, etc.
/// The query type must have a `bind` method that accepts `String`, `i64`, `f64`, and `bool`.
macro_rules! bind_sqipe_values {
    ($query:expr, $binds:expr) => {{
        let mut q = $query;
        for v in $binds {
            q = match v {
                sqipe::Value::String(s) => q.bind(s),
                sqipe::Value::Int(n) => q.bind(n),
                sqipe::Value::Float(f) => q.bind(f),
                sqipe::Value::Bool(b) => q.bind(b),
            };
        }
        q
    }};
}

pub(crate) use bind_sqipe_values;

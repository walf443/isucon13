/// MySQL で COUNT()、SUM() 等を使って DECIMAL 型の値になったものを i64 に変換するための構造体。
#[derive(Debug)]
pub struct MysqlDecimal(pub i64);
impl sqlx::Decode<'_, sqlx::MySql> for MysqlDecimal {
    fn decode(
        value: sqlx::mysql::MySqlValueRef,
    ) -> Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        use sqlx::{Type as _, ValueRef as _};

        let type_info = value.type_info();
        if i64::compatible(&type_info) {
            i64::decode(value).map(Self)
        } else if u64::compatible(&type_info) {
            let n = u64::decode(value)?.try_into()?;
            Ok(Self(n))
        } else if sqlx::types::Decimal::compatible(&type_info) {
            use num_traits::ToPrimitive as _;
            let n = sqlx::types::Decimal::decode(value)?
                .to_i64()
                .expect("failed to convert DECIMAL type to i64");
            Ok(Self(n))
        } else {
            todo!()
        }
    }
}
impl sqlx::Type<sqlx::MySql> for MysqlDecimal {
    fn type_info() -> sqlx::mysql::MySqlTypeInfo {
        i64::type_info()
    }

    fn compatible(ty: &sqlx::mysql::MySqlTypeInfo) -> bool {
        i64::compatible(ty) || u64::compatible(ty) || sqlx::types::Decimal::compatible(ty)
    }
}
impl From<MysqlDecimal> for i64 {
    fn from(value: MysqlDecimal) -> Self {
        value.0
    }
}

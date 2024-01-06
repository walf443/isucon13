use fake::{Dummy, Fake, Faker, Rng};
use serde::{Serialize, Serializer};
use sqlx::database::{HasArguments, HasValueRef};
use sqlx::encode::IsNull;
use sqlx::error::BoxDynError;
use sqlx::mysql::MySqlTypeInfo;
use sqlx::{Decode, Encode, MySql, Type};
use std::marker::PhantomData;

#[derive(Debug, sqlx::Type, PartialEq, Eq)]
pub struct Id<T> {
    id: i64,
    _phantom: PhantomData<T>,
}

impl<T> Serialize for Id<T> {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_i64(self.get())
    }
}

impl<T> Clone for Id<T> {
    fn clone(&self) -> Self {
        Self::new(self.get())
    }
}

impl<T> Id<T> {
    pub fn new(id: i64) -> Self {
        Self {
            id,
            _phantom: PhantomData,
        }
    }

    pub fn get(&self) -> i64 {
        self.id
    }
}

impl<T> Type<MySql> for Id<T> {
    fn type_info() -> MySqlTypeInfo {
        <i64 as Type<MySql>>::type_info()
    }

    fn compatible(ty: &MySqlTypeInfo) -> bool {
        <i64 as Type<MySql>>::compatible(ty)
    }
}

impl<T> Encode<'_, MySql> for Id<T> {
    fn encode_by_ref(&self, buf: &mut <MySql as HasArguments<'_>>::ArgumentBuffer) -> IsNull {
        <i64 as Encode<MySql>>::encode(self.get(), buf)
    }
}

impl<T> Decode<'_, MySql> for Id<T> {
    fn decode(value: <MySql as HasValueRef<'_>>::ValueRef) -> Result<Self, BoxDynError> {
        let val = <i64 as Decode<MySql>>::decode(value)?;
        Ok(Self::new(val))
    }
}

impl<T> Dummy<Faker> for Id<T> {
    fn dummy_with_rng<R: Rng + ?Sized>(config: &Faker, rng: &mut R) -> Self {
        let id = Fake::fake_with_rng::<i64, R>(config, rng);
        Self::new(id)
    }
}

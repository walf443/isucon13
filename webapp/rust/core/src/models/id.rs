use fake::{Dummy, Fake, Faker, Rng};
use serde::{Serialize, Serializer};
use sqlx::database::{HasArguments, HasValueRef};
use sqlx::encode::IsNull;
use sqlx::error::BoxDynError;
use sqlx::mysql::MySqlTypeInfo;
use sqlx::{Decode, Encode, MySql, Type};
use std::marker::PhantomData;

#[derive(Debug, sqlx::Type, PartialEq, Eq)]
pub struct Id<T, U: Clone> {
    id: U,
    _phantom: PhantomData<T>,
}

impl<T, U: Clone> Id<T, U> {
    pub fn new(id: U) -> Self {
        Self {
            id,
            _phantom: PhantomData,
        }
    }

    pub fn get(&self) -> U {
        self.id.clone()
    }
}

impl<T, U: Clone> From<U> for Id<T, U> {
    fn from(value: U) -> Self {
        Self::new(value)
    }
}

impl<T> Serialize for Id<T, i64> {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_i64(self.get())
    }
}

impl<T, U: Clone> Clone for Id<T, U> {
    fn clone(&self) -> Self {
        Self::new(self.get())
    }
}

impl<T, U: Clone + sqlx::Type<sqlx::MySql>> Type<MySql> for Id<T, U> {
    fn type_info() -> MySqlTypeInfo {
        <U as Type<MySql>>::type_info()
    }

    fn compatible(ty: &MySqlTypeInfo) -> bool {
        <U as Type<MySql>>::compatible(ty)
    }
}

impl<T, U: Clone + for<'a> sqlx::Encode<'a, sqlx::MySql>> Encode<'_, MySql> for Id<T, U> {
    fn encode_by_ref(&self, buf: &mut <MySql as HasArguments<'_>>::ArgumentBuffer) -> IsNull {
        <U as Encode<MySql>>::encode(self.get(), buf)
    }
}

impl<T, U: Clone + for<'a> sqlx::Decode<'a, sqlx::MySql>> Decode<'_, MySql> for Id<T, U> {
    fn decode(value: <MySql as HasValueRef<'_>>::ValueRef) -> Result<Self, BoxDynError> {
        let val = <U as Decode<MySql>>::decode(value)?;
        Ok(Self::new(val))
    }
}

impl<T, U: Clone + fake::Dummy<fake::Faker>> Dummy<Faker> for Id<T, U> {
    fn dummy_with_rng<R: Rng + ?Sized>(config: &Faker, rng: &mut R) -> Self {
        let id = Fake::fake_with_rng::<U, R>(config, rng);
        Self::new(id)
    }
}

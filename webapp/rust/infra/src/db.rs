use isupipe_core::db::DBPool;

/// 接続 URL からモデル登録済みの `toasty::Db` を組み立てる。
pub async fn build_db(url: &str, max_pool_size: usize) -> toasty::Result<DBPool> {
    toasty::Db::builder()
        .models(crate::tables::models())
        .max_pool_size(max_pool_size)
        .connect(url)
        .await
}

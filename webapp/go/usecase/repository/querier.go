package repository

// Querier は repository がクエリを実行する先 (DB またはトランザクション)。
// repository のメソッドはこれを引数で受け取ることで、トランザクションの内外どちらからでも呼べる。
// トランザクションの境界をコード上で明示し、repository を 1 メソッドずつテストしやすくするため、あえて引数で渡している。
//
// 中身は infra が決める不透明な値で、usecase は TxManager から受け取って repository に渡すだけで、中身を見ない
// (usecase が DB の実装 (GORM など) に依存しないため)。
type Querier interface {
	// querier は、Querier を満たせる型を限るための目印。小文字のメソッドなので、他のパッケージの型は、
	// QuerierBase を埋め込まない限り満たせない (関係のない値をうっかり渡すとコンパイルエラーになる)。
	querier()
}

// QuerierBase は Querier を実装する型 (infra の実装) に埋め込む。
type QuerierBase struct{}

func (QuerierBase) querier() {}

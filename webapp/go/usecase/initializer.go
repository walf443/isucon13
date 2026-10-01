package usecase

// Initializer はアプリケーションのデータ (DB・DNS) を初期状態に戻すためのインターフェース。
type Initializer interface {
	// Initialize は初期化を実行し、その出力を返す。失敗した場合も出力を返す。
	Initialize() ([]byte, error)
}

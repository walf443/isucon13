package usecase

// fakeInitializer はテストで設定した関数に処理を委ねる Initializer。
// メソッドが 1 つだけなのでインターフェースは埋め込まない。関数を設定せずに呼ぶと panic する。
type fakeInitializer struct {
	initialize func() ([]byte, error)
}

func (i *fakeInitializer) Initialize() ([]byte, error) {
	return i.initialize()
}

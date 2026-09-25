package usecase

// Logger は usecase からログを出力するためのインターフェース。echo.Logger が満たす。
type Logger interface {
	Infof(format string, args ...interface{})
}

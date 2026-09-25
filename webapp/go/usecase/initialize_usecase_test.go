package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestInitializeUsecase_Initialize(t *testing.T) {
	runs := 0
	initializer := &fakeInitializer{
		initialize: func() ([]byte, error) {
			runs++
			return []byte("+ mysql ...\n"), nil
		},
	}
	logger := &fakeLogger{}
	u := NewInitializeUsecase(initializer, logger)

	if err := u.Initialize(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs != 1 {
		t.Errorf("runs = %d, want 1", runs)
	}
	// 成功時はログを出さない (移行前と同じ)
	if len(logger.lines) != 0 || len(logger.warnLines) != 0 {
		t.Errorf("log lines = %q, warn lines = %q, want none", logger.lines, logger.warnLines)
	}
}

func TestInitializeUsecase_Initialize_Error(t *testing.T) {
	exitErr := errors.New("exit status 1")
	logger := &fakeLogger{}
	initializer := &fakeInitializer{
		initialize: func() ([]byte, error) { return []byte("ERROR 2003: Can't connect"), exitErr },
	}
	u := NewInitializeUsecase(initializer, logger)

	err := u.Initialize(context.Background())
	if !errors.Is(err, exitErr) {
		t.Fatalf("err = %v, want %v", err, exitErr)
	}
	// エラーには終了理由を、ログにはスクリプトの出力を載せる (移行前と同じ)
	if want := "failed to initialize: exit status 1"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
	if want := []string{"init.sh failed with err=ERROR 2003: Can't connect"}; !slices.Equal(logger.warnLines, want) {
		t.Errorf("warn lines = %q, want %q", logger.warnLines, want)
	}
}

package usecase

import (
	"context"
	"fmt"
)

type InitializeUsecase interface {
	// Initialize はアプリケーションのデータを初期状態に戻す。
	Initialize(ctx context.Context) error
}

type initializeUsecase struct {
	initializer Initializer
	logger      Logger
}

func NewInitializeUsecase(initializer Initializer, logger Logger) InitializeUsecase {
	return &initializeUsecase{initializer: initializer, logger: logger}
}

func (u *initializeUsecase) Initialize(ctx context.Context) error {
	if out, err := u.initializer.Initialize(); err != nil {
		// 移行前と同じく、ログにはスクリプトの出力を、エラーには終了理由を載せる
		u.logger.Warnf("init.sh failed with err=%s", string(out))
		return fmt.Errorf("failed to initialize: %w", err)
	}
	return nil
}

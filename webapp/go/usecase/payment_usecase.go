package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type PaymentUsecase interface {
	// FindTotalTip は全てのライブコメントのチップ合計を返す。
	FindTotalTip(ctx context.Context) (int64, error)
}

type paymentUsecase struct {
	txManager       repository.TxManager
	livecommentRepo repository.LivecommentRepository
}

func NewPaymentUsecase(txManager repository.TxManager, livecommentRepo repository.LivecommentRepository) PaymentUsecase {
	return &paymentUsecase{txManager: txManager, livecommentRepo: livecommentRepo}
}

func (u *paymentUsecase) FindTotalTip(ctx context.Context) (int64, error) {
	var totalTip int64
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		totalTip, err = u.livecommentRepo.SumTip(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to count total tip: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalTip, nil
}

package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestPaymentUsecase_FindTotalTip(t *testing.T) {
	repo := &fakeLivecommentRepository{
		sumTip: func(context.Context, repository.Querier) (int64, error) { return 1500, nil },
	}
	u := NewPaymentUsecase(&fakeTxManager{}, repo)

	got, err := u.FindTotalTip(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1500 {
		t.Errorf("total tip = %d, want 1500", got)
	}
}

func TestPaymentUsecase_FindTotalTip_Error(t *testing.T) {
	boom := errors.New("boom")
	repo := &fakeLivecommentRepository{
		sumTip: func(context.Context, repository.Querier) (int64, error) { return 0, boom },
	}
	u := NewPaymentUsecase(&fakeTxManager{}, repo)

	_, err := u.FindTotalTip(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if want := "failed to count total tip: boom"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

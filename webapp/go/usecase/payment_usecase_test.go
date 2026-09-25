package usecase

import (
	"context"
	"errors"
	"testing"
)

func TestPaymentUsecase_FindTotalTip(t *testing.T) {
	u := NewPaymentUsecase(&fakeTxManager{}, &fakeLivecommentRepository{totalTip: 1500})

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
	u := NewPaymentUsecase(&fakeTxManager{}, &fakeLivecommentRepository{totalTipErr: boom})

	_, err := u.FindTotalTip(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if want := "failed to count total tip: boom"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

func TestLivestreamViewerUsecase_Enter(t *testing.T) {
	repo := &fakeLivestreamViewersHistoryRepository{}
	u := NewLivestreamViewerUsecase(&fakeTxManager{}, repo).(*livestreamViewerUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	if err := u.Enter(context.Background(), 1, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := model.LivestreamViewersHistoryModel{UserID: 1, LivestreamID: 10, CreatedAt: 1700000000}
	if *repo.gotCreated != want {
		t.Errorf("created = %+v, want %+v", *repo.gotCreated, want)
	}
}

func TestLivestreamViewerUsecase_Enter_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewLivestreamViewerUsecase(&fakeTxManager{}, &fakeLivestreamViewersHistoryRepository{err: boom})

	err := u.Enter(context.Background(), 1, 10)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if want := "failed to insert livestream_view_history: boom"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestLivestreamViewerUsecase_Exit(t *testing.T) {
	repo := &fakeLivestreamViewersHistoryRepository{}
	u := NewLivestreamViewerUsecase(&fakeTxManager{}, repo)

	if err := u.Exit(context.Background(), 1, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.gotUserID != 1 || repo.gotLivestreamID != 10 {
		t.Errorf("userID = %d, livestreamID = %d, want 1, 10", repo.gotUserID, repo.gotLivestreamID)
	}
}

func TestLivestreamViewerUsecase_Exit_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewLivestreamViewerUsecase(&fakeTxManager{}, &fakeLivestreamViewersHistoryRepository{err: boom})

	err := u.Exit(context.Background(), 1, 10)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if want := "failed to delete livestream_view_history: boom"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

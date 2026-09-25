package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestLivestreamViewerUsecase_Enter(t *testing.T) {
	var created *model.LivestreamViewersHistoryModel
	repo := &fakeLivestreamViewersHistoryRepository{
		create: func(_ context.Context, _ repository.Querier, viewer *model.LivestreamViewersHistoryModel) error {
			created = viewer
			return nil
		},
	}
	u := NewLivestreamViewerUsecase(&fakeTxManager{}, repo).(*livestreamViewerUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	if err := u.Enter(context.Background(), 1, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := model.LivestreamViewersHistoryModel{UserID: 1, LivestreamID: 10, CreatedAt: 1700000000}
	if created == nil || *created != want {
		t.Errorf("created = %+v, want %+v", created, want)
	}
}

func TestLivestreamViewerUsecase_Enter_Error(t *testing.T) {
	boom := errors.New("boom")
	repo := &fakeLivestreamViewersHistoryRepository{
		create: func(context.Context, repository.Querier, *model.LivestreamViewersHistoryModel) error { return boom },
	}
	u := NewLivestreamViewerUsecase(&fakeTxManager{}, repo)

	err := u.Enter(context.Background(), 1, 10)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if want := "failed to insert livestream_view_history: boom"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestLivestreamViewerUsecase_Exit(t *testing.T) {
	repo := &fakeLivestreamViewersHistoryRepository{
		deleteByUserIDAndLivestreamID: func(_ context.Context, _ repository.Querier, userID model.UserID, livestreamID model.LivestreamID) error {
			if userID != 1 || livestreamID != 10 {
				t.Errorf("userID = %d, livestreamID = %d, want 1, 10", userID, livestreamID)
			}
			return nil
		},
	}
	u := NewLivestreamViewerUsecase(&fakeTxManager{}, repo)

	if err := u.Exit(context.Background(), 1, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLivestreamViewerUsecase_Exit_Error(t *testing.T) {
	boom := errors.New("boom")
	repo := &fakeLivestreamViewersHistoryRepository{
		deleteByUserIDAndLivestreamID: func(context.Context, repository.Querier, model.UserID, model.LivestreamID) error { return boom },
	}
	u := NewLivestreamViewerUsecase(&fakeTxManager{}, repo)

	err := u.Exit(context.Background(), 1, 10)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if want := "failed to delete livestream_view_history: boom"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

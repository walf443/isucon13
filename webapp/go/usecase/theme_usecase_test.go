package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestThemeUsecase_FindByUsername(t *testing.T) {
	themeRepo := &fakeThemeRepository{
		findByUserID: func(_ context.Context, _ repository.Querier, userID model.UserID) (*model.ThemeModel, error) {
			if userID != 1 {
				t.Errorf("userID = %d, want 1", userID)
			}
			return &model.ThemeModel{ID: 10, UserID: 1, DarkMode: true}, nil
		},
	}
	u := NewThemeUsecase(&fakeTxManager{}, newUserRepositoryFindingID(t, "alice", 1, nil), themeRepo)

	theme, err := u.FindByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if theme.ID != 10 || !theme.DarkMode {
		t.Errorf("theme = %+v", theme)
	}
}

func TestThemeUsecase_FindByUsername_UserNotFound(t *testing.T) {
	u := NewThemeUsecase(&fakeTxManager{}, newUserRepositoryFindingID(t, "alice", 0, repository.ErrNotFound), &fakeThemeRepository{})

	_, err := u.FindByUsername(context.Background(), "alice")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
}

func TestThemeUsecase_FindByUsername_ThemeNotFound(t *testing.T) {
	// テーマが無いのはデータ不整合なので、ユーザ不在 (404) とは区別する
	themeRepo := &fakeThemeRepository{
		findByUserID: func(context.Context, repository.Querier, model.UserID) (*model.ThemeModel, error) {
			return nil, repository.ErrNotFound
		},
	}
	u := NewThemeUsecase(&fakeTxManager{}, newUserRepositoryFindingID(t, "alice", 1, nil), themeRepo)

	_, err := u.FindByUsername(context.Background(), "alice")
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, ErrUserNotFound) {
		t.Errorf("err = %v, should not be ErrUserNotFound", err)
	}
}

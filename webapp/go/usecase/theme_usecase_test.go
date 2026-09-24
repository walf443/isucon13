package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestThemeUsecase_FindByUsername(t *testing.T) {
	themeRepo := &fakeThemeRepository{theme: &model.ThemeModel{ID: 10, UserID: 1, DarkMode: true}}
	u := NewThemeUsecase(&fakeTxManager{}, &fakeUserRepository{id: 1}, themeRepo)

	theme, err := u.FindByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if theme.ID != 10 || !theme.DarkMode {
		t.Errorf("theme = %+v", theme)
	}
	if themeRepo.gotUserID != 1 {
		t.Errorf("userID = %d, want 1", themeRepo.gotUserID)
	}
}

func TestThemeUsecase_FindByUsername_UserNotFound(t *testing.T) {
	u := NewThemeUsecase(&fakeTxManager{}, &fakeUserRepository{err: repository.ErrNotFound}, &fakeThemeRepository{})

	_, err := u.FindByUsername(context.Background(), "alice")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
}

func TestThemeUsecase_FindByUsername_ThemeNotFound(t *testing.T) {
	// テーマが無いのはデータ不整合なので、ユーザ不在 (404) とは区別する
	u := NewThemeUsecase(&fakeTxManager{}, &fakeUserRepository{id: 1}, &fakeThemeRepository{err: repository.ErrNotFound})

	_, err := u.FindByUsername(context.Background(), "alice")
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, ErrUserNotFound) {
		t.Errorf("err = %v, should not be ErrUserNotFound", err)
	}
}

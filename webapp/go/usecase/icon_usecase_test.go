package usecase

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestIconUsecase_FindImageByUsername(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		userRepo  *fakeUserRepository
		iconRepo  *fakeIconRepository
		wantImage []byte
		wantErr   error
	}{
		{
			name:      "returns registered icon",
			userRepo:  &fakeUserRepository{id: 1},
			iconRepo:  &fakeIconRepository{image: []byte("icon")},
			wantImage: []byte("icon"),
		},
		{
			name:     "user not found",
			userRepo: &fakeUserRepository{err: repository.ErrNotFound},
			iconRepo: &fakeIconRepository{},
			wantErr:  ErrUserNotFound,
		},
		{
			name:     "icon not registered",
			userRepo: &fakeUserRepository{id: 1},
			iconRepo: &fakeIconRepository{err: repository.ErrNotFound},
			wantErr:  ErrIconNotFound,
		},
		{
			name:     "icon repository error",
			userRepo: &fakeUserRepository{id: 1},
			iconRepo: &fakeIconRepository{err: boom},
			wantErr:  boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewIconUsecase(&fakeTxManager{}, tt.userRepo, tt.iconRepo)

			image, err := u.FindImageByUsername(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !bytes.Equal(image, tt.wantImage) {
				t.Errorf("image = %q, want %q", image, tt.wantImage)
			}
			if tt.wantErr == nil && tt.iconRepo.gotUserID != 1 {
				t.Errorf("userID = %d, want 1", tt.iconRepo.gotUserID)
			}
		})
	}
}

func TestIconUsecase_Update(t *testing.T) {
	iconRepo := &fakeIconRepository{createID: 100}
	u := NewIconUsecase(&fakeTxManager{}, &fakeUserRepository{}, iconRepo)

	iconID, err := u.Update(context.Background(), 1, []byte("new icon"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if iconID != 100 {
		t.Errorf("iconID = %d, want 100", iconID)
	}
	// 古いアイコンを消してから登録する
	if want := []string{"DeleteByUserID", "Create"}; !slices.Equal(iconRepo.calls, want) {
		t.Errorf("calls = %v, want %v", iconRepo.calls, want)
	}
	if iconRepo.gotUserID != 1 || string(iconRepo.gotImage) != "new icon" {
		t.Errorf("userID = %d, image = %q", iconRepo.gotUserID, iconRepo.gotImage)
	}
}

func TestIconUsecase_Update_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		iconRepo  *fakeIconRepository
		wantCalls []string
	}{
		{
			name:      "delete fails",
			iconRepo:  &fakeIconRepository{deleteErr: boom},
			wantCalls: []string{"DeleteByUserID"},
		},
		{
			name:      "create fails",
			iconRepo:  &fakeIconRepository{createErr: boom},
			wantCalls: []string{"DeleteByUserID", "Create"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewIconUsecase(&fakeTxManager{}, &fakeUserRepository{}, tt.iconRepo)

			_, err := u.Update(context.Background(), 1, []byte("new icon"))
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
			if !slices.Equal(tt.iconRepo.calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", tt.iconRepo.calls, tt.wantCalls)
			}
		})
	}
}

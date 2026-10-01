package usecase

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestIconUsecase_FindImageByUsername(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// userID, userErr はユーザ名から ID を引いた結果
		userID  domain.UserID
		userErr error
		// iconImage, iconErr はアイコンの取得が返す値
		iconImage []byte
		iconErr   error
		wantImage []byte
		wantErr   error
	}{
		{
			name:      "returns registered icon",
			userID:    1,
			iconImage: []byte("icon"),
			wantImage: []byte("icon"),
		},
		{
			name:    "user not found",
			userErr: repository.ErrNotFound,
			wantErr: ErrUserNotFound,
		},
		{
			name:    "icon not registered",
			userID:  1,
			iconErr: repository.ErrNotFound,
			wantErr: ErrIconNotFound,
		},
		{
			name:    "icon repository error",
			userID:  1,
			iconErr: boom,
			wantErr: boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iconRepo := &fakeIconRepository{
				findImageByUserID: func(_ context.Context, _ repository.Querier, userID domain.UserID) ([]byte, error) {
					if userID != 1 {
						t.Errorf("userID = %d, want 1", userID)
					}
					return tt.iconImage, tt.iconErr
				},
			}
			u := NewIconUsecase(&fakeTxManager{}, newUserRepositoryFindingID(t, "alice", tt.userID, tt.userErr), iconRepo)

			image, err := u.FindImageByUsername(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !bytes.Equal(image, tt.wantImage) {
				t.Errorf("image = %q, want %q", image, tt.wantImage)
			}
		})
	}
}

// newIconRepositoryForUpdate は Update で呼ばれるメソッドを、呼ばれた順に calls へ記録する fakeIconRepository を返す。
func newIconRepositoryForUpdate(t *testing.T, calls *[]string, deleteErr, createErr error) *fakeIconRepository {
	return &fakeIconRepository{
		deleteByUserID: func(_ context.Context, _ repository.Querier, userID domain.UserID) error {
			*calls = append(*calls, "DeleteByUserID")
			if userID != 1 {
				t.Errorf("delete userID = %d, want 1", userID)
			}
			return deleteErr
		},
		create: func(_ context.Context, _ repository.Querier, userID domain.UserID, image []byte) (domain.IconID, error) {
			*calls = append(*calls, "Create")
			if userID != 1 || string(image) != "new icon" {
				t.Errorf("create userID = %d, image = %q", userID, image)
			}
			return 100, createErr
		},
	}
}

func TestIconUsecase_Update(t *testing.T) {
	var calls []string
	u := NewIconUsecase(&fakeTxManager{}, &fakeUserRepository{}, newIconRepositoryForUpdate(t, &calls, nil, nil))

	iconID, err := u.Update(context.Background(), 1, []byte("new icon"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if iconID != 100 {
		t.Errorf("iconID = %d, want 100", iconID)
	}
	// 古いアイコンを消してから登録する
	if want := []string{"DeleteByUserID", "Create"}; !slices.Equal(calls, want) {
		t.Errorf("calls = %v, want %v", calls, want)
	}
}

func TestIconUsecase_Update_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		deleteErr error
		createErr error
		wantCalls []string
	}{
		{
			name:      "delete fails",
			deleteErr: boom,
			wantCalls: []string{"DeleteByUserID"},
		},
		{
			name:      "create fails",
			createErr: boom,
			wantCalls: []string{"DeleteByUserID", "Create"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			u := NewIconUsecase(&fakeTxManager{}, &fakeUserRepository{}, newIconRepositoryForUpdate(t, &calls, tt.deleteErr, tt.createErr))

			_, err := u.Update(context.Background(), 1, []byte("new icon"))
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
			if !slices.Equal(calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", calls, tt.wantCalls)
			}
		})
	}
}

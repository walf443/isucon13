package usecase

import (
	"context"
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestUserFiller_Fill(t *testing.T) {
	alice := &domain.User{ID: 1, Name: "alice", DisplayName: "Alice", Description: "hi"}
	bob := &domain.User{ID: 2, Name: "bob"}
	themes := map[domain.UserID]*domain.Theme{
		1: {ID: 10, UserID: 1, DarkMode: true},
		2: {ID: 20, UserID: 2},
	}
	icons := map[domain.UserID][]byte{1: []byte("icon")}
	var themeCalls []domain.UserID
	f := newUserFillerForTest(themes, icons, &themeCalls)

	// 同じユーザが重複していても 1 回だけ取得する
	got, err := f.Fill(context.Background(), nil, []*domain.User{alice, bob, alice})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.UserID]domain.UserDetail{
		1: {ID: 1, Name: "alice", DisplayName: "Alice", Description: "hi", Theme: *themes[1], IconHash: domain.HashIcon([]byte("icon"))},
		// アイコン未登録のユーザは既定のハッシュ
		2: {ID: 2, Name: "bob", Theme: *themes[2], IconHash: "default-hash"},
	}
	if len(got) != len(want) {
		t.Fatalf("len(users) = %d, want %d", len(got), len(want))
	}
	for id, w := range want {
		if u, ok := got[id]; !ok || *u != w {
			t.Errorf("users[%d] = %+v, want %+v", id, u, w)
		}
	}
	if want := []domain.UserID{1, 2}; !slices.Equal(themeCalls, want) {
		t.Errorf("theme calls = %v, want %v", themeCalls, want)
	}
}

func TestUserFiller_Fill_Empty(t *testing.T) {
	got, err := newUserFillerForTest(nil, nil, nil).Fill(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("users = %v, want empty", slices.Collect(maps.Keys(got)))
	}
}

func TestUserFiller_Fill_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		themeRepo *fakeThemeRepository
		iconRepo  *fakeIconRepository
		wantErr   error
		wantMsg   string
	}{
		{
			name: "theme not found",
			themeRepo: &fakeThemeRepository{
				findByUserID: func(context.Context, repository.Querier, domain.UserID) (*domain.Theme, error) {
					return nil, repository.ErrNotFound
				},
			},
			wantErr: errMissingDetail,
			wantMsg: "failed to get theme of user 1: not found",
		},
		{
			name: "get icon fails",
			themeRepo: &fakeThemeRepository{
				findByUserID: func(context.Context, repository.Querier, domain.UserID) (*domain.Theme, error) {
					return &domain.Theme{ID: 10, UserID: 1}, nil
				},
			},
			iconRepo: &fakeIconRepository{
				findImageByUserID: func(context.Context, repository.Querier, domain.UserID) ([]byte, error) { return nil, boom },
			},
			wantErr: boom,
			wantMsg: "failed to get icon of user 1: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewUserFiller(tt.themeRepo, tt.iconRepo, "")
			_, err := f.Fill(context.Background(), nil, []*domain.User{{ID: 1}})
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == errMissingDetail && errors.Is(err, repository.ErrNotFound)) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

package usecase

import (
	"context"
	"errors"
	"maps"
	"reflect"
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
	var themeCalls [][]domain.UserID
	f := newUserFillerForTest(themes, icons, &themeCalls)

	// 同じユーザが重複していても、重複を除いた ID の一覧で 1 回だけ取得する
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
	if want := [][]domain.UserID{{1, 2}}; !reflect.DeepEqual(themeCalls, want) {
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
	themeOf1 := func(context.Context, repository.Querier, []domain.UserID) ([]*domain.Theme, error) {
		return []*domain.Theme{{ID: 10, UserID: 1}}, nil
	}
	noIcons := func(context.Context, repository.Querier, []domain.UserID) ([]*domain.Icon, error) { return nil, nil }

	tests := []struct {
		name      string
		themeRepo *fakeThemeRepository
		iconRepo  *fakeIconRepository
		wantErr   error
		wantMsg   string
	}{
		{
			// ユーザ 2 のテーマが無い
			name:      "theme not found",
			themeRepo: &fakeThemeRepository{findAllByUserIDs: themeOf1},
			iconRepo:  &fakeIconRepository{findAllByUserIDs: noIcons},
			wantErr:   errMissingDetail,
			wantMsg:   "failed to get theme of user 2: not found",
		},
		{
			name: "get themes fails",
			themeRepo: &fakeThemeRepository{
				findAllByUserIDs: func(context.Context, repository.Querier, []domain.UserID) ([]*domain.Theme, error) { return nil, boom },
			},
			wantErr: boom,
			wantMsg: "failed to get themes: boom",
		},
		{
			name:      "get icons fails",
			themeRepo: &fakeThemeRepository{findAllByUserIDs: themeOf1},
			iconRepo: &fakeIconRepository{
				findAllByUserIDs: func(context.Context, repository.Querier, []domain.UserID) ([]*domain.Icon, error) { return nil, boom },
			},
			wantErr: boom,
			wantMsg: "failed to get icons: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewUserFiller(tt.themeRepo, tt.iconRepo, "")
			_, err := f.Fill(context.Background(), nil, []*domain.User{{ID: 1}, {ID: 2}})
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == errMissingDetail && errors.Is(err, repository.ErrNotFound)) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

package usecase

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestAsMissingDetail(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		err  error
		// wantMissing は errMissingDetail として判定されることを期待するかどうか
		wantMissing bool
		// wantErr は errors.Is で判定できることを期待する元のエラー (無ければ nil)
		wantErr error
	}{
		{name: "not found", err: repository.ErrNotFound, wantMissing: true},
		{name: "wrapped not found", err: fmt.Errorf("wrapped: %w", repository.ErrNotFound), wantMissing: true},
		{name: "other error", err: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := asMissingDetail(tt.err)
			// メッセージは変えない
			if got.Error() != tt.err.Error() {
				t.Errorf("message = %q, want %q", got.Error(), tt.err.Error())
			}
			if errors.Is(got, errMissingDetail) != tt.wantMissing {
				t.Errorf("errors.Is(errMissingDetail) = %v, want %v", !tt.wantMissing, tt.wantMissing)
			}
			// 付随するデータの欠損は ErrNotFound (主となるデータの不在) として判定されない
			if tt.wantMissing && errors.Is(got, repository.ErrNotFound) {
				t.Errorf("err = %v, should not be ErrNotFound", got)
			}
			if tt.wantErr != nil && !errors.Is(got, tt.wantErr) {
				t.Errorf("err = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestUniqueKeys(t *testing.T) {
	users := []*domain.User{{ID: 2}, {ID: 1}, {ID: 2}, {ID: 3}, {ID: 1}}
	got := uniqueKeys(users, func(u *domain.User) domain.UserID { return u.ID })
	// 重複を除き、最初に現れた順
	if want := []domain.UserID{2, 1, 3}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got := uniqueKeys([]*domain.User(nil), func(u *domain.User) domain.UserID { return u.ID }); got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}
}

func TestIndexBy(t *testing.T) {
	themes := []*domain.Theme{{ID: 10, UserID: 1}, {ID: 20, UserID: 2}}
	got := indexBy(themes, func(theme *domain.Theme) domain.UserID { return theme.UserID })
	if len(got) != 2 || got[1] != themes[0] || got[2] != themes[1] {
		t.Errorf("got %v", got)
	}
}

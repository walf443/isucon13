package mysql

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestThemeRepository_FindByUserID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	lastID := insertSQL(t, tx, "INSERT INTO themes (user_id, dark_mode) VALUES (?, ?)", 42, true)
	wantID := domain.ThemeID(lastID)

	theme, err := NewThemeRepository().FindByUserID(ctx, tx, 42)
	if err != nil {
		t.Fatalf("FindByUserID returned error: %v", err)
	}
	if theme.ID != wantID || theme.UserID != 42 || !theme.DarkMode {
		t.Errorf("theme = %+v, want ID=%d UserID=42 DarkMode=true", theme, wantID)
	}
}

func TestThemeRepository_FindByUserID_NotFound(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewThemeRepository().FindByUserID(context.Background(), tx, 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestThemeRepository_FindAllByUserIDs(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	alice := insertTestUser(t, tx, "alice")
	bob := insertTestUser(t, tx, "bob")
	// carol はテーマが無い
	carol := insertTestUser(t, tx, "carol")
	repo := NewThemeRepository()
	for _, theme := range []*domain.Theme{{UserID: alice, DarkMode: true}, {UserID: bob}} {
		if err := repo.Create(ctx, tx, theme); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	got, err := repo.FindAllByUserIDs(ctx, tx, []domain.UserID{alice, bob, carol})
	if err != nil {
		t.Fatalf("FindAllByUserIDs returned error: %v", err)
	}
	darkModes := map[domain.UserID]bool{}
	for _, theme := range got {
		darkModes[theme.UserID] = theme.DarkMode
	}
	if want := map[domain.UserID]bool{alice: true, bob: false}; !maps.Equal(darkModes, want) {
		t.Errorf("dark modes = %v, want %v", darkModes, want)
	}

	got, err = repo.FindAllByUserIDs(ctx, tx, nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("FindAllByUserIDs(nil) = %#v, %v, want empty", got, err)
	}
}

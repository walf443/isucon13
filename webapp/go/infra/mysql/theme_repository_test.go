package mysql

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestThemeRepository_FindByUserID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	res, err := tx.ExecContext(ctx, "INSERT INTO themes (user_id, dark_mode) VALUES (?, ?)", 42, true)
	if err != nil {
		t.Fatalf("failed to insert theme: %v", err)
	}
	lastID, _ := res.LastInsertId()
	wantID := model.ThemeID(lastID)

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

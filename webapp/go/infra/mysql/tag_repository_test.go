package mysql

import (
	"context"
	"testing"
)

func TestTagRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	for _, name := range []string{"ライブ配信", "ゲーム実況"} {
		if _, err := tx.ExecContext(ctx, "INSERT INTO tags (name) VALUES (?)", name); err != nil {
			t.Fatalf("failed to insert tag: %v", err)
		}
	}

	tags, err := NewTagRepository().FindAll(ctx, tx)
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}

	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
	got := map[string]bool{}
	for _, tag := range tags {
		if tag.ID == 0 {
			t.Errorf("tag %q has zero ID", tag.Name)
		}
		got[tag.Name] = true
	}
	for _, want := range []string{"ライブ配信", "ゲーム実況"} {
		if !got[want] {
			t.Errorf("tag %q not found in %v", want, got)
		}
	}
}

func TestTagRepository_FindAll_Empty(t *testing.T) {
	tx := beginTestTx(t)

	tags, err := NewTagRepository().FindAll(context.Background(), tx)
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("len(tags) = %d, want 0", len(tags))
	}
}

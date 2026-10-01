package mysql

import (
	"context"
	"maps"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
)

func TestTagRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	for _, name := range []string{"ライブ配信", "ゲーム実況"} {
		execSQL(t, tx, "INSERT INTO tags (name) VALUES (?)", name)
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

func TestTagRepository_FindIDsByName(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	lastID := insertSQL(t, tx, "INSERT INTO tags (name) VALUES (?)", "ゲーム実況")
	wantID := domain.TagID(lastID)
	repo := NewTagRepository()

	ids, err := repo.FindIDsByName(ctx, tx, "ゲーム実況")
	if err != nil {
		t.Fatalf("FindIDsByName returned error: %v", err)
	}
	if len(ids) != 1 || ids[0] != wantID {
		t.Errorf("ids = %v, want [%d]", ids, wantID)
	}

	ids, err = repo.FindIDsByName(ctx, tx, "存在しないタグ")
	if err != nil {
		t.Fatalf("FindIDsByName returned error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("ids = %v, want empty", ids)
	}
}

func TestTagRepository_FindAllByIDs(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	var ids []domain.TagID
	for _, name := range []string{"ゲーム実況", "雑談", "歌枠"} {
		id := insertSQL(t, tx, "INSERT INTO tags (name) VALUES (?)", name)
		ids = append(ids, domain.TagID(id))
	}
	repo := NewTagRepository()

	got, err := repo.FindAllByIDs(ctx, tx, []domain.TagID{ids[2], ids[0], 999999})
	if err != nil {
		t.Fatalf("FindAllByIDs returned error: %v", err)
	}
	names := map[domain.TagID]string{}
	for _, tag := range got {
		names[tag.ID] = tag.Name
	}
	if want := map[domain.TagID]string{ids[0]: "ゲーム実況", ids[2]: "歌枠"}; !maps.Equal(names, want) {
		t.Errorf("tags = %v, want %v", names, want)
	}

	got, err = repo.FindAllByIDs(ctx, tx, nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("FindAllByIDs(nil) = %#v, %v, want empty", got, err)
	}
}

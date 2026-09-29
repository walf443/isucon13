package mysql

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func insertTestLivestream(t *testing.T, tx repository.Querier, userID domain.UserID, title string) domain.LivestreamID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(),
		"INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, title, "desc "+title, "https://example.com/"+title+".m3u8", "https://example.com/"+title+".jpg", 1700000000, 1700003600)
	if err != nil {
		t.Fatalf("failed to insert livestream: %v", err)
	}
	id, _ := res.LastInsertId()
	return domain.LivestreamID(id)
}

func insertTestTag(t *testing.T, tx repository.Querier, livestreamID domain.LivestreamID, name string) domain.Tag {
	t.Helper()
	ctx := context.Background()
	res, err := tx.ExecContext(ctx, "INSERT INTO tags (name) VALUES (?)", name)
	if err != nil {
		t.Fatalf("failed to insert tag: %v", err)
	}
	tagID, _ := res.LastInsertId()
	if _, err := tx.ExecContext(ctx, "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)", livestreamID, tagID); err != nil {
		t.Fatalf("failed to insert livestream_tag: %v", err)
	}
	return domain.Tag{ID: domain.TagID(tagID), Name: name}
}

func TestLivestreamRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	repo := NewLivestreamRepository()

	got, err := repo.FindByID(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	want := domain.Livestream{
		ID:           livestreamID,
		UserID:       ownerID,
		Title:        "stream",
		Description:  "desc stream",
		PlaylistUrl:  "https://example.com/stream.m3u8",
		ThumbnailUrl: "https://example.com/stream.jpg",
		StartAt:      1700000000,
		EndAt:        1700003600,
	}
	if *got != want {
		t.Errorf("got %+v, want %+v", *got, want)
	}

	if _, err := repo.FindByID(ctx, tx, 999999); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestLivestreamRepository_FindAllByIDAndUserID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	otherID := insertTestUser(t, tx, "bob")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	repo := NewLivestreamRepository()

	got, err := repo.FindAllByIDAndUserID(ctx, tx, livestreamID, ownerID)
	if err != nil {
		t.Fatalf("FindAllByIDAndUserID returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != livestreamID || got[0].UserID != ownerID {
		t.Errorf("got %+v, want the livestream %d", got, livestreamID)
	}

	for _, tt := range []struct {
		name   string
		id     domain.LivestreamID
		userID domain.UserID
	}{
		{name: "other user", id: livestreamID, userID: otherID},
		{name: "livestream not found", id: 999999, userID: ownerID},
	} {
		got, err := repo.FindAllByIDAndUserID(ctx, tx, tt.id, tt.userID)
		if err != nil {
			t.Fatalf("%s: FindAllByIDAndUserID returned error: %v", tt.name, err)
		}
		if len(got) != 0 {
			t.Errorf("%s: got %+v, want empty", tt.name, got)
		}
	}
}

func TestLivestreamRepository_CreateAndLivestreamTagRepository_Create(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewLivestreamRepository()
	livestreamTagRepo := NewLivestreamTagRepository()

	ownerID := insertTestUser(t, tx, "alice")
	var tagIDs []domain.TagID
	for _, name := range []string{"tag-a", "tag-b"} {
		res, err := tx.ExecContext(ctx, "INSERT INTO tags (name) VALUES (?)", name)
		if err != nil {
			t.Fatalf("failed to insert tag: %v", err)
		}
		id, _ := res.LastInsertId()
		tagIDs = append(tagIDs, domain.TagID(id))
	}

	want := domain.Livestream{
		UserID:       ownerID,
		Title:        "stream",
		Description:  "desc",
		PlaylistUrl:  "https://example.com/p.m3u8",
		ThumbnailUrl: "https://example.com/t.jpg",
		StartAt:      1700874000,
		EndAt:        1700877600,
	}
	id, err := repo.Create(ctx, tx, &want)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	want.ID = id
	for _, tagID := range tagIDs {
		if err := livestreamTagRepo.Create(ctx, tx, id, tagID); err != nil {
			t.Fatalf("livestream tag Create returned error: %v", err)
		}
	}

	got, err := repo.FindByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if *got != want {
		t.Errorf("livestream = %+v, want %+v", *got, want)
	}
	livestreamTags, err := livestreamTagRepo.FindAllByLivestreamID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindAllByLivestreamID returned error: %v", err)
	}
	gotTagIDs := make([]domain.TagID, len(livestreamTags))
	for i, lt := range livestreamTags {
		gotTagIDs[i] = lt.TagID
	}
	slices.Sort(gotTagIDs)
	if !slices.Equal(gotTagIDs, tagIDs) {
		t.Errorf("tag IDs = %v, want %v", gotTagIDs, tagIDs)
	}
}

func livestreamIDs(livestreams []*domain.Livestream) []domain.LivestreamID {
	ids := make([]domain.LivestreamID, len(livestreams))
	for i, l := range livestreams {
		ids[i] = l.ID
	}
	return ids
}

func TestLivestreamRepository_FindAllOrderByIDDesc(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	first := insertTestLivestream(t, tx, ownerID, "first")
	second := insertTestLivestream(t, tx, ownerID, "second")
	third := insertTestLivestream(t, tx, ownerID, "third")
	repo := NewLivestreamRepository()

	got, err := repo.FindAllOrderByIDDesc(ctx, tx)
	if err != nil {
		t.Fatalf("FindAllOrderByIDDesc returned error: %v", err)
	}
	if ids, want := livestreamIDs(got), []domain.LivestreamID{third, second, first}; !slices.Equal(ids, want) {
		t.Errorf("IDs = %v, want %v", ids, want)
	}

	got, err = repo.FindAllOrderByIDDescLimited(ctx, tx, 2)
	if err != nil {
		t.Fatalf("FindAllOrderByIDDescLimited returned error: %v", err)
	}
	if ids, want := livestreamIDs(got), []domain.LivestreamID{third, second}; !slices.Equal(ids, want) {
		t.Errorf("IDs = %v, want %v", ids, want)
	}
}

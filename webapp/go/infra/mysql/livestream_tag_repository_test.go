package mysql

import (
	"context"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
)

func livestreamTagLivestreamIDs(livestreamTags []*domain.LivestreamTag) []domain.LivestreamID {
	ids := make([]domain.LivestreamID, len(livestreamTags))
	for i, lt := range livestreamTags {
		ids[i] = lt.LivestreamID
	}
	return ids
}

func TestLivestreamTagRepository_FindAllByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	otherID := insertTestLivestream(t, tx, ownerID, "other")
	tag1 := insertTestTag(t, tx, livestreamID, "ゲーム実況")
	tag2 := insertTestTag(t, tx, livestreamID, "雑談")
	insertTestTag(t, tx, otherID, "歌枠")
	repo := NewLivestreamTagRepository()

	got, err := repo.FindAllByLivestreamID(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindAllByLivestreamID returned error: %v", err)
	}
	tagIDs := make([]domain.TagID, len(got))
	for i, lt := range got {
		if lt.ID == 0 || lt.LivestreamID != livestreamID {
			t.Errorf("livestream tag = %+v", lt)
		}
		tagIDs[i] = lt.TagID
	}
	slices.Sort(tagIDs)
	if want := []domain.TagID{tag1.ID, tag2.ID}; !slices.Equal(tagIDs, want) {
		t.Errorf("tag IDs = %v, want %v", tagIDs, want)
	}

	got, err = repo.FindAllByLivestreamID(ctx, tx, 999999)
	if err != nil {
		t.Fatalf("FindAllByLivestreamID returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("livestream tags = %v, want empty", got)
	}
}

func TestLivestreamTagRepository_FindAllByTagIDs(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	first := insertTestLivestream(t, tx, ownerID, "first")
	second := insertTestLivestream(t, tx, ownerID, "second")
	third := insertTestLivestream(t, tx, ownerID, "third")
	tagA := insertTestTag(t, tx, first, "A")
	tagB := insertTestTag(t, tx, third, "B")
	insertTestTag(t, tx, second, "C")

	got, err := NewLivestreamTagRepository().FindAllByTagIDs(ctx, tx, []domain.TagID{tagA.ID, tagB.ID})
	if err != nil {
		t.Fatalf("FindAllByTagIDs returned error: %v", err)
	}
	// ライブ配信の ID の降順
	if ids, want := livestreamTagLivestreamIDs(got), []domain.LivestreamID{third, first}; !slices.Equal(ids, want) {
		t.Errorf("livestream IDs = %v, want %v", ids, want)
	}
}

func TestLivestreamTagRepository_FindAllByLivestreamIDs(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	first := insertTestLivestream(t, tx, ownerID, "first")
	second := insertTestLivestream(t, tx, ownerID, "second")
	other := insertTestLivestream(t, tx, ownerID, "other")
	tagA := insertTestTag(t, tx, second, "A")
	tagB := insertTestTag(t, tx, first, "B")
	tagC := insertTestTag(t, tx, second, "C")
	insertTestTag(t, tx, other, "D")
	repo := NewLivestreamTagRepository()

	got, err := repo.FindAllByLivestreamIDs(ctx, tx, []domain.LivestreamID{first, second})
	if err != nil {
		t.Fatalf("FindAllByLivestreamIDs returned error: %v", err)
	}
	// 紐付けの ID の昇順 (登録した順)
	tagIDs := make([]domain.TagID, len(got))
	for i, lt := range got {
		tagIDs[i] = lt.TagID
	}
	if want := []domain.TagID{tagA.ID, tagB.ID, tagC.ID}; !slices.Equal(tagIDs, want) {
		t.Errorf("tag IDs = %v, want %v", tagIDs, want)
	}

	got, err = repo.FindAllByLivestreamIDs(ctx, tx, nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("FindAllByLivestreamIDs(nil) = %#v, %v, want empty", got, err)
	}
}

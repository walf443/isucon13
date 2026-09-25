package mysql

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func reactionIDs(reactions []*model.Reaction) []int64 {
	ids := make([]int64, len(reactions))
	for i, r := range reactions {
		ids[i] = r.ID
	}
	return ids
}

func TestReactionRepository_CreateAndFindWithDetailsByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, ownerID, true)
	viewerID := insertTestUser(t, tx, "bob")
	insertTestTheme(t, tx, viewerID, false)
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	repo := NewReactionRepository(nil)

	id, err := repo.Create(ctx, tx, &model.ReactionModel{UserID: viewerID, LivestreamID: livestreamID, EmojiName: "tada", CreatedAt: 1700000000})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := repo.FindWithDetailsByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindWithDetailsByID returned error: %v", err)
	}
	if got.ID != id || got.EmojiName != "tada" || got.CreatedAt != 1700000000 {
		t.Errorf("reaction = %+v", got)
	}
	if got.User.ID != viewerID || got.User.Name != "bob" {
		t.Errorf("user = %+v", got.User)
	}
	if got.Livestream.ID != livestreamID || got.Livestream.Owner.ID != ownerID {
		t.Errorf("livestream = %+v", got.Livestream)
	}
}

func TestReactionRepository_FindWithDetailsByID_NotFound(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewReactionRepository(nil).FindWithDetailsByID(context.Background(), tx, 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestReactionRepository_FindWithDetailsByID_LivestreamNotFound(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	userID := insertTestUser(t, tx, "bob")
	insertTestTheme(t, tx, userID, false)
	repo := NewReactionRepository(nil)
	id, err := repo.Create(ctx, tx, &model.ReactionModel{UserID: userID, LivestreamID: 999999, EmojiName: "tada", CreatedAt: 1})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// ライブ配信の欠損はデータ不整合なので、リアクション不在 (ErrNotFound) とは区別する
	_, err = repo.FindWithDetailsByID(ctx, tx, id)
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, should not be ErrNotFound", err)
	}
}

func TestReactionRepository_FindAllWithDetailsByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	userID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, userID, false)
	livestreamID := insertTestLivestream(t, tx, userID, "stream")
	otherLivestreamID := insertTestLivestream(t, tx, userID, "other")
	repo := NewReactionRepository(nil)

	create := func(livestreamID int64, createdAt int64) int64 {
		t.Helper()
		id, err := repo.Create(ctx, tx, &model.ReactionModel{UserID: userID, LivestreamID: livestreamID, EmojiName: "tada", CreatedAt: createdAt})
		if err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
		return id
	}
	// 作成日時の降順になることを確認するため、ID の順序とはずらす
	r1 := create(livestreamID, 300)
	r2 := create(livestreamID, 100)
	r3 := create(livestreamID, 200)
	create(otherLivestreamID, 400)

	all, err := repo.FindAllWithDetailsByLivestreamID(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindAllWithDetailsByLivestreamID returned error: %v", err)
	}
	if want := []int64{r1, r3, r2}; !slices.Equal(reactionIDs(all), want) {
		t.Errorf("ids = %v, want %v", reactionIDs(all), want)
	}

	limited, err := repo.FindAllWithDetailsByLivestreamIDLimited(ctx, tx, livestreamID, 2)
	if err != nil {
		t.Fatalf("FindAllWithDetailsByLivestreamIDLimited returned error: %v", err)
	}
	if want := []int64{r1, r3}; !slices.Equal(reactionIDs(limited), want) {
		t.Errorf("ids = %v, want %v", reactionIDs(limited), want)
	}
}

package mysql

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func reactionModelIDs(reactions []*domain.ReactionModel) []domain.ReactionID {
	ids := make([]domain.ReactionID, len(reactions))
	for i, r := range reactions {
		ids[i] = r.ID
	}
	return ids
}

func TestReactionRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	repo := NewReactionRepository()
	want := domain.ReactionModel{UserID: ownerID, LivestreamID: livestreamID, EmojiName: "tada", CreatedAt: 1700000000}
	id, err := repo.Create(ctx, tx, &want)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	want.ID = id

	got, err := repo.FindByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if *got != want {
		t.Errorf("reaction = %+v, want %+v", *got, want)
	}

	if _, err := repo.FindByID(ctx, tx, 999999); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestReactionRepository_FindAllByLivestreamIDOrderByCreatedAtDesc(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	otherID := insertTestLivestream(t, tx, ownerID, "other")
	repo := NewReactionRepository()
	create := func(livestreamID domain.LivestreamID, createdAt int64) domain.ReactionID {
		id, err := repo.Create(ctx, tx, &domain.ReactionModel{UserID: ownerID, LivestreamID: livestreamID, EmojiName: "tada", CreatedAt: createdAt})
		if err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
		return id
	}
	oldest := create(livestreamID, 100)
	newest := create(livestreamID, 300)
	middle := create(livestreamID, 200)
	create(otherID, 400)

	got, err := repo.FindAllByLivestreamIDOrderByCreatedAtDesc(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindAllByLivestreamIDOrderByCreatedAtDesc returned error: %v", err)
	}
	if ids, want := reactionModelIDs(got), []domain.ReactionID{newest, middle, oldest}; !slices.Equal(ids, want) {
		t.Errorf("IDs = %v, want %v", ids, want)
	}

	got, err = repo.FindAllByLivestreamIDOrderByCreatedAtDescLimited(ctx, tx, livestreamID, 2)
	if err != nil {
		t.Fatalf("FindAllByLivestreamIDOrderByCreatedAtDescLimited returned error: %v", err)
	}
	if ids, want := reactionModelIDs(got), []domain.ReactionID{newest, middle}; !slices.Equal(ids, want) {
		t.Errorf("IDs = %v, want %v", ids, want)
	}
}

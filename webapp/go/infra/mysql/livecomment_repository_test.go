package mysql

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func insertTestLivecomment(t *testing.T, tx repository.Querier, userID domain.UserID, livestreamID domain.LivestreamID, comment string, createdAt int64) domain.LivecommentID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(), "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)", userID, livestreamID, comment, 10, createdAt)
	if err != nil {
		t.Fatalf("failed to insert livecomment: %v", err)
	}
	id, _ := res.LastInsertId()
	return domain.LivecommentID(id)
}

func TestLivecommentRepository_Create(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	viewerID := insertTestUser(t, tx, "bob")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	repo := NewLivecommentRepository()

	want := domain.LivecommentModel{UserID: viewerID, LivestreamID: livestreamID, Comment: "hello", Tip: 500, CreatedAt: 1700000000}
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
		t.Errorf("livecomment = %+v, want %+v", *got, want)
	}
}

func TestLivecommentRepository_DeleteAllByLivestreamIDMatchingNGWord(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	viewerID := insertTestUser(t, tx, "bob")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	otherLivestreamID := insertTestLivestream(t, tx, ownerID, "other")

	hit := insertTestLivecomment(t, tx, viewerID, livestreamID, "this is bad", 100)
	// LIKE による判定なので大文字小文字は区別しない (移行前と同じ)
	hitUpper := insertTestLivecomment(t, tx, viewerID, livestreamID, "BAD!", 100)
	safe := insertTestLivecomment(t, tx, viewerID, livestreamID, "good", 100)
	// 他のライブ配信のライブコメントは消さない
	otherLivestream := insertTestLivecomment(t, tx, viewerID, otherLivestreamID, "bad", 100)

	if err := NewLivecommentRepository().DeleteAllByLivestreamIDMatchingNGWord(ctx, tx, livestreamID, "bad"); err != nil {
		t.Fatalf("DeleteAllByLivestreamIDMatchingNGWord returned error: %v", err)
	}

	var remaining []domain.LivecommentID
	if err := tx.SelectContext(ctx, &remaining, "SELECT id FROM livecomments ORDER BY id"); err != nil {
		t.Fatalf("failed to get livecomments: %v", err)
	}
	if want := []domain.LivecommentID{safe, otherLivestream}; !slices.Equal(remaining, want) {
		t.Errorf("remaining = %v, want %v (deleted should be %v, %v)", remaining, want, hit, hitUpper)
	}
}

func TestLivecommentRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	id := insertTestLivecomment(t, tx, ownerID, livestreamID, "hello", 100)
	repo := NewLivecommentRepository()

	got, err := repo.FindByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if want := (domain.LivecommentModel{ID: id, UserID: ownerID, LivestreamID: livestreamID, Comment: "hello", Tip: 10, CreatedAt: 100}); *got != want {
		t.Errorf("got %+v, want %+v", *got, want)
	}

	if _, err := repo.FindByID(ctx, tx, 999999); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestLivecommentRepository_SumTip(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewLivecommentRepository()

	before, err := repo.SumTip(ctx, tx)
	if err != nil {
		t.Fatalf("SumTip returned error: %v", err)
	}

	ownerID := insertTestUser(t, tx, "alice")
	stream := insertTestLivestream(t, tx, ownerID, "s1")
	otherStream := insertTestLivestream(t, tx, ownerID, "s2")
	insertTestLivecommentWithTip(t, tx, ownerID, stream, 100)
	insertTestLivecommentWithTip(t, tx, ownerID, otherStream, 250)

	// ライブ配信によらず全てのライブコメントのチップを合計する
	after, err := repo.SumTip(ctx, tx)
	if err != nil {
		t.Fatalf("SumTip returned error: %v", err)
	}
	if after-before != 350 {
		t.Errorf("total tip = %d (before %d), want +350", after, before)
	}
}

func livecommentModelIDs(livecomments []*domain.LivecommentModel) []domain.LivecommentID {
	ids := make([]domain.LivecommentID, len(livecomments))
	for i, l := range livecomments {
		ids[i] = l.ID
	}
	return ids
}

func TestLivecommentRepository_FindAllByLivestreamIDOrderByCreatedAtDesc(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	otherID := insertTestLivestream(t, tx, ownerID, "other")
	oldest := insertTestLivecomment(t, tx, ownerID, livestreamID, "oldest", 100)
	newest := insertTestLivecomment(t, tx, ownerID, livestreamID, "newest", 300)
	middle := insertTestLivecomment(t, tx, ownerID, livestreamID, "middle", 200)
	insertTestLivecomment(t, tx, ownerID, otherID, "other", 400)
	repo := NewLivecommentRepository()

	got, err := repo.FindAllByLivestreamIDOrderByCreatedAtDesc(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindAllByLivestreamIDOrderByCreatedAtDesc returned error: %v", err)
	}
	if ids, want := livecommentModelIDs(got), []domain.LivecommentID{newest, middle, oldest}; !slices.Equal(ids, want) {
		t.Errorf("IDs = %v, want %v", ids, want)
	}

	got, err = repo.FindAllByLivestreamIDOrderByCreatedAtDescLimited(ctx, tx, livestreamID, 2)
	if err != nil {
		t.Fatalf("FindAllByLivestreamIDOrderByCreatedAtDescLimited returned error: %v", err)
	}
	if ids, want := livecommentModelIDs(got), []domain.LivecommentID{newest, middle}; !slices.Equal(ids, want) {
		t.Errorf("IDs = %v, want %v", ids, want)
	}
}

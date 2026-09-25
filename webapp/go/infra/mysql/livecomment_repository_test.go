package mysql

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
	"github.com/jmoiron/sqlx"
)

func insertTestLivecomment(t *testing.T, tx *sqlx.Tx, userID model.UserID, livestreamID model.LivestreamID, comment string, createdAt int64) model.LivecommentID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(), "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)", userID, livestreamID, comment, 10, createdAt)
	if err != nil {
		t.Fatalf("failed to insert livecomment: %v", err)
	}
	id, _ := res.LastInsertId()
	return model.LivecommentID(id)
}

func livecommentIDs(livecomments []*model.Livecomment) []model.LivecommentID {
	ids := make([]model.LivecommentID, len(livecomments))
	for i, l := range livecomments {
		ids[i] = l.ID
	}
	return ids
}

func TestLivecommentRepository_FindAllWithDetailsByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, ownerID, true)
	viewerID := insertTestUser(t, tx, "bob")
	insertTestTheme(t, tx, viewerID, false)
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	otherLivestreamID := insertTestLivestream(t, tx, ownerID, "other")

	// 作成日時の降順になることを確認するため、ID の順序とはずらす
	c1 := insertTestLivecomment(t, tx, viewerID, livestreamID, "c1", 300)
	c2 := insertTestLivecomment(t, tx, viewerID, livestreamID, "c2", 100)
	c3 := insertTestLivecomment(t, tx, viewerID, livestreamID, "c3", 200)
	insertTestLivecomment(t, tx, viewerID, otherLivestreamID, "other", 400)
	repo := NewLivecommentRepository(nil)

	all, err := repo.FindAllWithDetailsByLivestreamID(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindAllWithDetailsByLivestreamID returned error: %v", err)
	}
	if want := []model.LivecommentID{c1, c3, c2}; !slices.Equal(livecommentIDs(all), want) {
		t.Errorf("ids = %v, want %v", livecommentIDs(all), want)
	}
	got := all[0]
	if got.Comment != "c1" || got.Tip != 10 || got.CreatedAt != 300 || got.User.ID != viewerID || got.Livestream.ID != livestreamID || got.Livestream.Owner.ID != ownerID {
		t.Errorf("livecomment = %+v", got)
	}

	limited, err := repo.FindAllWithDetailsByLivestreamIDLimited(ctx, tx, livestreamID, 2)
	if err != nil {
		t.Fatalf("FindAllWithDetailsByLivestreamIDLimited returned error: %v", err)
	}
	if want := []model.LivecommentID{c1, c3}; !slices.Equal(livecommentIDs(limited), want) {
		t.Errorf("ids = %v, want %v", livecommentIDs(limited), want)
	}
}

func TestLivecommentReportRepository_FindAllWithDetailsByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, ownerID, true)
	viewerID := insertTestUser(t, tx, "bob")
	insertTestTheme(t, tx, viewerID, false)
	reporterID := insertTestUser(t, tx, "carol")
	insertTestTheme(t, tx, reporterID, false)
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	otherLivestreamID := insertTestLivestream(t, tx, ownerID, "other")
	commentID := insertTestLivecomment(t, tx, viewerID, livestreamID, "bad comment", 100)
	otherCommentID := insertTestLivecomment(t, tx, viewerID, otherLivestreamID, "other", 100)

	insertReport := func(livestreamID model.LivestreamID, livecommentID model.LivecommentID) model.LivecommentReportID {
		t.Helper()
		res, err := tx.ExecContext(ctx, "INSERT INTO livecomment_reports (user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)", reporterID, livestreamID, livecommentID, 200)
		if err != nil {
			t.Fatalf("failed to insert livecomment report: %v", err)
		}
		id, _ := res.LastInsertId()
		return model.LivecommentReportID(id)
	}
	reportID := insertReport(livestreamID, commentID)
	insertReport(otherLivestreamID, otherCommentID)

	reports, err := NewLivecommentReportRepository(nil).FindAllWithDetailsByLivestreamID(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindAllWithDetailsByLivestreamID returned error: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("len(reports) = %d, want 1", len(reports))
	}
	got := reports[0]
	if got.ID != reportID || got.CreatedAt != 200 || got.Reporter.ID != reporterID {
		t.Errorf("report = %+v", got)
	}
	if got.Livecomment.ID != commentID || got.Livecomment.User.ID != viewerID || got.Livecomment.Livestream.ID != livestreamID {
		t.Errorf("livecomment = %+v", got.Livecomment)
	}
}

func TestLivecommentReportRepository_LivecommentNotFound(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	reporterID := insertTestUser(t, tx, "carol")
	insertTestTheme(t, tx, reporterID, false)
	if _, err := tx.ExecContext(ctx, "INSERT INTO livecomment_reports (user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)", reporterID, 1, 999999, 200); err != nil {
		t.Fatalf("failed to insert livecomment report: %v", err)
	}

	// 報告されたライブコメントの欠損はデータ不整合なので ErrNotFound にはしない
	_, err := NewLivecommentReportRepository(nil).FindAllWithDetailsByLivestreamID(ctx, tx, 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, should not be ErrNotFound", err)
	}
}

func TestLivecommentRepository_CreateAndFindWithDetailsByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, ownerID, true)
	viewerID := insertTestUser(t, tx, "bob")
	insertTestTheme(t, tx, viewerID, false)
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	repo := NewLivecommentRepository(nil)

	id, err := repo.Create(ctx, tx, &model.LivecommentModel{UserID: viewerID, LivestreamID: livestreamID, Comment: "hello", Tip: 500, CreatedAt: 1700000000})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := repo.FindWithDetailsByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindWithDetailsByID returned error: %v", err)
	}
	if got.ID != id || got.Comment != "hello" || got.Tip != 500 || got.CreatedAt != 1700000000 {
		t.Errorf("livecomment = %+v", got)
	}
	if got.User.ID != viewerID || got.Livestream.ID != livestreamID || got.Livestream.Owner.ID != ownerID {
		t.Errorf("user = %+v, livestream = %+v", got.User, got.Livestream)
	}

	if _, err := repo.FindWithDetailsByID(ctx, tx, 999999); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
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

	if err := NewLivecommentRepository(nil).DeleteAllByLivestreamIDMatchingNGWord(ctx, tx, livestreamID, "bad"); err != nil {
		t.Fatalf("DeleteAllByLivestreamIDMatchingNGWord returned error: %v", err)
	}

	var remaining []model.LivecommentID
	if err := tx.SelectContext(ctx, &remaining, "SELECT id FROM livecomments ORDER BY id"); err != nil {
		t.Fatalf("failed to get livecomments: %v", err)
	}
	if want := []model.LivecommentID{safe, otherLivestream}; !slices.Equal(remaining, want) {
		t.Errorf("remaining = %v, want %v (deleted should be %v, %v)", remaining, want, hit, hitUpper)
	}
}

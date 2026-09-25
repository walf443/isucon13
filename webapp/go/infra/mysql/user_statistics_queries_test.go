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

// ユーザ統計情報 (GET /api/user/:username/statistics) で使う集計系のクエリのテスト

func insertTestReaction(t *testing.T, tx *sqlx.Tx, userID model.UserID, livestreamID model.LivestreamID, emojiName string) {
	t.Helper()
	if _, err := NewReactionRepository(nil).Create(context.Background(), tx, &model.ReactionModel{UserID: userID, LivestreamID: livestreamID, EmojiName: emojiName, CreatedAt: 100}); err != nil {
		t.Fatalf("failed to insert reaction: %v", err)
	}
}

func insertTestLivecommentWithTip(t *testing.T, tx *sqlx.Tx, userID model.UserID, livestreamID model.LivestreamID, tip int64) model.LivecommentID {
	t.Helper()
	id, err := NewLivecommentRepository(nil).Create(context.Background(), tx, &model.LivecommentModel{UserID: userID, LivestreamID: livestreamID, Comment: "c", Tip: tip, CreatedAt: 100})
	if err != nil {
		t.Fatalf("failed to insert livecomment: %v", err)
	}
	return id
}

func TestUserRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	aliceID := insertTestUser(t, tx, "alice")
	bobID := insertTestUser(t, tx, "bob")

	users, err := NewUserRepository(nil).FindAll(ctx, tx)
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	got := map[model.UserID]model.UserModel{}
	for _, u := range users {
		got[u.ID] = *u
	}
	if want := (model.UserModel{ID: aliceID, Name: "alice", DisplayName: "Display alice", Description: "desc alice", HashedPassword: "hashed"}); got[aliceID] != want {
		t.Errorf("alice = %+v, want %+v", got[aliceID], want)
	}
	if got[bobID].Name != "bob" {
		t.Errorf("bob = %+v", got[bobID])
	}
}

func TestReactionRepository_CountByLivestreamOwner(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewReactionRepository(nil)

	ownerID := insertTestUser(t, tx, "alice")
	otherOwnerID := insertTestUser(t, tx, "bob")
	viewerID := insertTestUser(t, tx, "carol")
	stream1 := insertTestLivestream(t, tx, ownerID, "s1")
	stream2 := insertTestLivestream(t, tx, ownerID, "s2")
	otherStream := insertTestLivestream(t, tx, otherOwnerID, "other")

	insertTestReaction(t, tx, viewerID, stream1, "smile")
	insertTestReaction(t, tx, viewerID, stream1, "heart")
	insertTestReaction(t, tx, viewerID, stream2, "smile")
	// 他の配信者のライブ配信へのリアクションは数えない
	insertTestReaction(t, tx, viewerID, otherStream, "smile")
	// 配信者自身のリアクションかどうかは関係ない (配信者のライブ配信へのリアクションを数える)
	insertTestReaction(t, tx, otherOwnerID, stream2, "cry")

	byID, err := repo.CountByLivestreamOwnerID(ctx, tx, ownerID)
	if err != nil {
		t.Fatalf("CountByLivestreamOwnerID returned error: %v", err)
	}
	byName, err := repo.CountByLivestreamOwnerName(ctx, tx, "alice")
	if err != nil {
		t.Fatalf("CountByLivestreamOwnerName returned error: %v", err)
	}
	if byID != 4 || byName != 4 {
		t.Errorf("byID = %d, byName = %d, want 4", byID, byName)
	}

	// ライブ配信やリアクションが無いユーザは 0
	if n, err := repo.CountByLivestreamOwnerID(ctx, tx, viewerID); err != nil || n != 0 {
		t.Errorf("CountByLivestreamOwnerID(viewer) = %d, %v, want 0", n, err)
	}
	if n, err := repo.CountByLivestreamOwnerName(ctx, tx, "unknown"); err != nil || n != 0 {
		t.Errorf("CountByLivestreamOwnerName(unknown) = %d, %v, want 0", n, err)
	}
}

func TestReactionRepository_FindFavoriteEmojiByLivestreamOwnerName(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewReactionRepository(nil)

	ownerID := insertTestUser(t, tx, "alice")
	viewerID := insertTestUser(t, tx, "bob")
	insertTestUser(t, tx, "carol")
	stream := insertTestLivestream(t, tx, ownerID, "s1")

	for _, emoji := range []string{"apple", "apple", "zebra", "zebra", "smile"} {
		insertTestReaction(t, tx, viewerID, stream, emoji)
	}

	// 同数の場合は絵文字名の降順で先頭のもの (移行前と同じ)
	got, err := repo.FindFavoriteEmojiByLivestreamOwnerName(ctx, tx, "alice")
	if err != nil {
		t.Fatalf("FindFavoriteEmojiByLivestreamOwnerName returned error: %v", err)
	}
	if got != "zebra" {
		t.Errorf("favorite emoji = %q, want zebra", got)
	}

	// リアクションが無い場合は ErrNotFound
	if _, err := repo.FindFavoriteEmojiByLivestreamOwnerName(ctx, tx, "carol"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestLivecommentRepository_SumTipByLivestreamOwnerID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewLivecommentRepository(nil)

	ownerID := insertTestUser(t, tx, "alice")
	otherOwnerID := insertTestUser(t, tx, "bob")
	viewerID := insertTestUser(t, tx, "carol")
	stream1 := insertTestLivestream(t, tx, ownerID, "s1")
	stream2 := insertTestLivestream(t, tx, ownerID, "s2")
	otherStream := insertTestLivestream(t, tx, otherOwnerID, "other")

	insertTestLivecommentWithTip(t, tx, viewerID, stream1, 100)
	insertTestLivecommentWithTip(t, tx, viewerID, stream2, 50)
	insertTestLivecommentWithTip(t, tx, viewerID, stream2, 0)
	insertTestLivecommentWithTip(t, tx, viewerID, otherStream, 1000)

	got, err := repo.SumTipByLivestreamOwnerID(ctx, tx, ownerID)
	if err != nil {
		t.Fatalf("SumTipByLivestreamOwnerID returned error: %v", err)
	}
	if got != 150 {
		t.Errorf("tips = %d, want 150", got)
	}

	// ライブコメントが無い場合は (NULL ではなく) 0
	if n, err := repo.SumTipByLivestreamOwnerID(ctx, tx, viewerID); err != nil || n != 0 {
		t.Errorf("SumTipByLivestreamOwnerID(viewer) = %d, %v, want 0", n, err)
	}
}

func TestLivecommentRepository_FindAllByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	stream := insertTestLivestream(t, tx, ownerID, "s1")
	otherStream := insertTestLivestream(t, tx, ownerID, "s2")
	c1 := insertTestLivecommentWithTip(t, tx, ownerID, stream, 100)
	c2 := insertTestLivecommentWithTip(t, tx, ownerID, stream, 50)
	insertTestLivecommentWithTip(t, tx, ownerID, otherStream, 10)

	livecomments, err := NewLivecommentRepository(nil).FindAllByLivestreamID(ctx, tx, stream)
	if err != nil {
		t.Fatalf("FindAllByLivestreamID returned error: %v", err)
	}
	// ORDER BY が無いので順序には依存しない
	var ids []model.LivecommentID
	for _, l := range livecomments {
		ids = append(ids, l.ID)
		if l.ID == c1 {
			if want := (model.LivecommentModel{ID: c1, UserID: ownerID, LivestreamID: stream, Comment: "c", Tip: 100, CreatedAt: 100}); *l != want {
				t.Errorf("livecomment = %+v, want %+v", *l, want)
			}
		}
	}
	slices.Sort(ids)
	if want := []model.LivecommentID{c1, c2}; !slices.Equal(ids, want) {
		t.Errorf("ids = %v, want %v", ids, want)
	}
}

func TestLivestreamRepository_FindAllByUserID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	otherID := insertTestUser(t, tx, "bob")
	s1 := insertTestLivestream(t, tx, ownerID, "s1")
	s2 := insertTestLivestream(t, tx, ownerID, "s2")
	insertTestLivestream(t, tx, otherID, "other")

	livestreams, err := NewLivestreamRepository(nil).FindAllByUserID(ctx, tx, ownerID)
	if err != nil {
		t.Fatalf("FindAllByUserID returned error: %v", err)
	}
	var ids []model.LivestreamID
	for _, l := range livestreams {
		ids = append(ids, l.ID)
	}
	slices.Sort(ids)
	if want := []model.LivestreamID{s1, s2}; !slices.Equal(ids, want) {
		t.Errorf("ids = %v, want %v", ids, want)
	}

	if got, err := NewLivestreamRepository(nil).FindAllByUserID(ctx, tx, 999999); err != nil || len(got) != 0 {
		t.Errorf("FindAllByUserID(unknown) = %v, %v, want empty", got, err)
	}
}

func TestLivestreamViewersHistoryRepository_CountByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewLivestreamViewersHistoryRepository()

	for _, v := range []model.LivestreamViewersHistoryModel{
		{UserID: 1, LivestreamID: 10, CreatedAt: 100},
		{UserID: 1, LivestreamID: 10, CreatedAt: 200},
		{UserID: 2, LivestreamID: 10, CreatedAt: 100},
		{UserID: 1, LivestreamID: 20, CreatedAt: 100},
	} {
		if err := repo.Create(ctx, tx, &v); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	// 同じユーザの重複した視聴履歴もそのまま数える (移行前と同じ)
	if n, err := repo.CountByLivestreamID(ctx, tx, 10); err != nil || n != 3 {
		t.Errorf("CountByLivestreamID(10) = %d, %v, want 3", n, err)
	}
	if n, err := repo.CountByLivestreamID(ctx, tx, 30); err != nil || n != 0 {
		t.Errorf("CountByLivestreamID(30) = %d, %v, want 0", n, err)
	}
}

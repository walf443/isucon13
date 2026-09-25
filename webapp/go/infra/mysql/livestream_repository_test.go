package mysql

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
	"github.com/jmoiron/sqlx"
)

func insertTestLivestream(t *testing.T, tx *sqlx.Tx, userID model.UserID, title string) model.LivestreamID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(),
		"INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, title, "desc "+title, "https://example.com/"+title+".m3u8", "https://example.com/"+title+".jpg", 1700000000, 1700003600)
	if err != nil {
		t.Fatalf("failed to insert livestream: %v", err)
	}
	id, _ := res.LastInsertId()
	return model.LivestreamID(id)
}

func insertTestTag(t *testing.T, tx *sqlx.Tx, livestreamID model.LivestreamID, name string) model.TagModel {
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
	return model.TagModel{ID: model.TagID(tagID), Name: name}
}

func TestLivestreamRepository_FindWithDetailsByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	fallback := []byte("fallback")

	ownerID := insertTestUser(t, tx, "alice")
	themeID := insertTestTheme(t, tx, ownerID, true)
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream1")
	tag1 := insertTestTag(t, tx, livestreamID, "ゲーム実況")
	tag2 := insertTestTag(t, tx, livestreamID, "雑談")

	got, err := NewLivestreamRepository(fallback).FindWithDetailsByID(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindWithDetailsByID returned error: %v", err)
	}

	want := &model.Livestream{
		ID: livestreamID,
		Owner: model.User{
			ID:          ownerID,
			Name:        "alice",
			DisplayName: "Display alice",
			Description: "desc alice",
			Theme:       model.ThemeModel{ID: themeID, UserID: ownerID, DarkMode: true},
			IconHash:    model.IconHash(fallback),
		},
		Title:        "stream1",
		Description:  "desc stream1",
		PlaylistUrl:  "https://example.com/stream1.m3u8",
		ThumbnailUrl: "https://example.com/stream1.jpg",
		Tags:         []model.TagModel{tag1, tag2},
		StartAt:      1700000000,
		EndAt:        1700003600,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

func TestLivestreamRepository_FindWithDetailsByID_NoTags(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, ownerID, false)
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream1")

	got, err := NewLivestreamRepository(nil).FindWithDetailsByID(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindWithDetailsByID returned error: %v", err)
	}
	if len(got.Tags) != 0 {
		t.Errorf("tags = %+v, want empty", got.Tags)
	}
}

func TestLivestreamRepository_FindWithDetailsByID_NotFound(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewLivestreamRepository(nil).FindWithDetailsByID(context.Background(), tx, 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestLivestreamRepository_FindWithDetailsByID_OwnerNotFound(t *testing.T) {
	tx := beginTestTx(t)
	livestreamID := insertTestLivestream(t, tx, 999999, "orphan")

	// 配信者の欠損はデータ不整合なので、ライブ配信不在 (ErrNotFound) とは区別する
	_, err := NewLivestreamRepository(nil).FindWithDetailsByID(context.Background(), tx, livestreamID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, should not be ErrNotFound", err)
	}
}

func TestFillLivestreams(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	aliceID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, aliceID, false)
	bobID := insertTestUser(t, tx, "bob")
	insertTestTheme(t, tx, bobID, false)

	aliceStreamID := insertTestLivestream(t, tx, aliceID, "alice-stream")
	aliceTag := insertTestTag(t, tx, aliceStreamID, "alice-tag")
	bobStreamID := insertTestLivestream(t, tx, bobID, "bob-stream")

	livestreamModels := []*model.LivestreamModel{
		// 入力の順序が保たれることを確認するため、ID の降順で渡す
		{ID: bobStreamID, UserID: bobID, Title: "bob-stream"},
		{ID: aliceStreamID, UserID: aliceID, Title: "alice-stream"},
	}

	livestreams, err := fillLivestreams(ctx, tx, livestreamModels, nil)
	if err != nil {
		t.Fatalf("fillLivestreams returned error: %v", err)
	}
	if len(livestreams) != 2 {
		t.Fatalf("len(livestreams) = %d, want 2", len(livestreams))
	}
	if livestreams[0].ID != bobStreamID || livestreams[0].Owner.ID != bobID || len(livestreams[0].Tags) != 0 {
		t.Errorf("livestreams[0] = %+v", livestreams[0])
	}
	if livestreams[1].ID != aliceStreamID || livestreams[1].Owner.ID != aliceID || !reflect.DeepEqual(livestreams[1].Tags, []model.TagModel{aliceTag}) {
		t.Errorf("livestreams[1] = %+v", livestreams[1])
	}
}

func TestLivestreamRepository_FindAllWithDetailsByUserID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	aliceID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, aliceID, false)
	bobID := insertTestUser(t, tx, "bob")
	insertTestTheme(t, tx, bobID, false)

	stream1 := insertTestLivestream(t, tx, aliceID, "alice-1")
	tag := insertTestTag(t, tx, stream1, "alice-tag")
	stream2 := insertTestLivestream(t, tx, aliceID, "alice-2")
	insertTestLivestream(t, tx, bobID, "bob-1")

	repo := NewLivestreamRepository(nil)

	got, err := repo.FindAllWithDetailsByUserID(ctx, tx, aliceID)
	if err != nil {
		t.Fatalf("FindAllWithDetailsByUserID returned error: %v", err)
	}
	// ORDER BY が無いので順序には依存しない
	gotByID := map[model.LivestreamID]*model.Livestream{}
	for _, l := range got {
		gotByID[l.ID] = l
	}
	if len(got) != 2 || gotByID[stream1] == nil || gotByID[stream2] == nil {
		t.Fatalf("got livestreams %+v, want IDs %d and %d", got, stream1, stream2)
	}
	if gotByID[stream1].Owner.ID != aliceID || !reflect.DeepEqual(gotByID[stream1].Tags, []model.TagModel{tag}) {
		t.Errorf("stream1 = %+v", gotByID[stream1])
	}

	// 配信が無いユーザは空
	carolID := insertTestUser(t, tx, "carol")
	got, err = repo.FindAllWithDetailsByUserID(ctx, tx, carolID)
	if err != nil {
		t.Fatalf("FindAllWithDetailsByUserID returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}

func livestreamIDs(livestreams []*model.Livestream) []model.LivestreamID {
	ids := make([]model.LivestreamID, len(livestreams))
	for i, l := range livestreams {
		ids[i] = l.ID
	}
	return ids
}

func TestLivestreamRepository_FindAllWithDetailsByTagIDs(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, ownerID, false)
	stream1 := insertTestLivestream(t, tx, ownerID, "s1")
	stream2 := insertTestLivestream(t, tx, ownerID, "s2")
	stream3 := insertTestLivestream(t, tx, ownerID, "s3")
	gameTag := insertTestTag(t, tx, stream1, "ゲーム実況")
	if _, err := tx.ExecContext(ctx, "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)", stream3, gameTag.ID); err != nil {
		t.Fatalf("failed to insert livestream_tag: %v", err)
	}
	insertTestTag(t, tx, stream2, "雑談")

	got, err := NewLivestreamRepository(nil).FindAllWithDetailsByTagIDs(ctx, tx, []model.TagID{gameTag.ID})
	if err != nil {
		t.Fatalf("FindAllWithDetailsByTagIDs returned error: %v", err)
	}
	// livestream_id の降順
	if want := []model.LivestreamID{stream3, stream1}; !slices.Equal(livestreamIDs(got), want) {
		t.Errorf("ids = %v, want %v", livestreamIDs(got), want)
	}
	if got[1].Owner.ID != ownerID || !reflect.DeepEqual(got[1].Tags, []model.TagModel{gameTag}) {
		t.Errorf("got[1] = %+v", got[1])
	}
}

func TestLivestreamRepository_FindAllWithDetails(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	insertTestTheme(t, tx, ownerID, false)
	stream1 := insertTestLivestream(t, tx, ownerID, "s1")
	stream2 := insertTestLivestream(t, tx, ownerID, "s2")
	stream3 := insertTestLivestream(t, tx, ownerID, "s3")
	repo := NewLivestreamRepository(nil)

	all, err := repo.FindAllWithDetails(ctx, tx)
	if err != nil {
		t.Fatalf("FindAllWithDetails returned error: %v", err)
	}
	// id の降順
	if want := []model.LivestreamID{stream3, stream2, stream1}; !slices.Equal(livestreamIDs(all), want) {
		t.Errorf("ids = %v, want %v", livestreamIDs(all), want)
	}

	limited, err := repo.FindAllWithDetailsLimited(ctx, tx, 2)
	if err != nil {
		t.Fatalf("FindAllWithDetailsLimited returned error: %v", err)
	}
	if want := []model.LivestreamID{stream3, stream2}; !slices.Equal(livestreamIDs(limited), want) {
		t.Errorf("ids = %v, want %v", livestreamIDs(limited), want)
	}
}

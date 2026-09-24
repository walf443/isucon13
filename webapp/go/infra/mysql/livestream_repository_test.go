package mysql

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
	"github.com/jmoiron/sqlx"
)

func insertTestLivestream(t *testing.T, tx *sqlx.Tx, userID int64, title string) int64 {
	t.Helper()
	res, err := tx.ExecContext(context.Background(),
		"INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, title, "desc "+title, "https://example.com/"+title+".m3u8", "https://example.com/"+title+".jpg", 1700000000, 1700003600)
	if err != nil {
		t.Fatalf("failed to insert livestream: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func insertTestTag(t *testing.T, tx *sqlx.Tx, livestreamID int64, name string) model.TagModel {
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
	return model.TagModel{ID: tagID, Name: name}
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

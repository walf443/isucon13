package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// find に渡る ID は重複を除いて昇順に並び、その順に分割されること。
func TestFindInChunked_SendsSortedDistinctIDsInChunks(t *testing.T) {
	var chunks [][]int
	rows, err := findInChunked([]int{3, 1, 3, 5, 2, 1, 4}, 2, func(chunk []int) ([]*int, error) {
		chunks = append(chunks, slices.Clone(chunk))
		found := make([]*int, len(chunk))
		for i := range chunk {
			found[i] = &chunk[i]
		}
		return found, nil
	})
	if err != nil {
		t.Fatalf("findInChunked returned error: %v", err)
	}

	if want := [][]int{{1, 2}, {3, 4}, {5}}; !reflect.DeepEqual(chunks, want) {
		t.Errorf("chunks = %v, want %v", chunks, want)
	}
	// 分割した結果は、渡した順に連結される
	got := make([]int, len(rows))
	for i, r := range rows {
		got[i] = *r
	}
	if want := []int{1, 2, 3, 4, 5}; !slices.Equal(got, want) {
		t.Errorf("rows = %v, want %v", got, want)
	}
}

func TestFindInChunked_EmptyDoesNotCallFind(t *testing.T) {
	rows, err := findInChunked([]int(nil), 2, func([]int) ([]*int, error) {
		t.Error("find should not be called")
		return nil, nil
	})
	if err != nil || rows == nil || len(rows) != 0 {
		t.Errorf("findInChunked(nil) = %#v, %v, want empty non-nil slice", rows, err)
	}
}

func TestFindInChunked_ReturnsFindError(t *testing.T) {
	boom := errors.New("boom")
	calls := 0
	_, err := findInChunked([]int{1, 2, 3}, 1, func([]int) ([]*int, error) {
		calls++
		return nil, boom
	})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
	// 最初のエラーで打ち切る
	if calls != 1 {
		t.Errorf("find calls = %d, want 1", calls)
	}
}

// withManyIDs は real の ID のあいだに、存在しない ID を水増しして、MySQL のプレースホルダーの上限 (65535) を超える数にした一覧を返す。
func withManyIDs[ID ~int64](real ...ID) []ID {
	ids := []ID{real[0]}
	for id := ID(1_000_000); len(ids) < 70_000; id++ {
		ids = append(ids, id)
	}
	return append(ids, real[1:]...)
}

func int64s[ID ~int64](ids []ID) []int64 {
	converted := make([]int64, len(ids))
	for i, id := range ids {
		converted[i] = int64(id)
	}
	return converted
}

// GORM に移行した repository の一括取得のメソッドが、MySQL のプレースホルダーの上限を超える数の ID でも、IN ? を分割して引けること。
// repository を GORM に移行するたびに、その一括取得のメソッドをこの表に足す (findIn を使っていないと、ここで失敗する)。
func TestMigratedRepositories_BulkMethodsAcceptMoreIDsThanMySQLPlaceholderLimit(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		// run はデータを登録し、一括取得のメソッドを水増しした ID で呼んで、期待する行の ID と返ってきた行の ID を返す。
		run func(t *testing.T, tx repository.Querier) (want, got []int64)
	}{
		{
			name: "UserRepository.FindAllByIDs",
			run: func(t *testing.T, tx repository.Querier) ([]int64, []int64) {
				alice, bob := insertTestUser(t, tx, "alice"), insertTestUser(t, tx, "bob")
				users, err := NewUserRepository().FindAllByIDs(ctx, tx, withManyIDs(alice, bob))
				if err != nil {
					t.Fatalf("FindAllByIDs returned error: %v", err)
				}
				got := make([]int64, len(users))
				for i, u := range users {
					got[i] = int64(u.ID)
				}
				return int64s([]domain.UserID{alice, bob}), got
			},
		},
		{
			name: "ThemeRepository.FindAllByUserIDs",
			run: func(t *testing.T, tx repository.Querier) ([]int64, []int64) {
				alice, bob := insertTestUser(t, tx, "alice"), insertTestUser(t, tx, "bob")
				for _, userID := range []domain.UserID{alice, bob} {
					if err := NewThemeRepository().Create(ctx, tx, &domain.Theme{UserID: userID}); err != nil {
						t.Fatalf("Create returned error: %v", err)
					}
				}
				themes, err := NewThemeRepository().FindAllByUserIDs(ctx, tx, withManyIDs(alice, bob))
				if err != nil {
					t.Fatalf("FindAllByUserIDs returned error: %v", err)
				}
				got := make([]int64, len(themes))
				for i, theme := range themes {
					got[i] = int64(theme.UserID)
				}
				return int64s([]domain.UserID{alice, bob}), got
			},
		},
		{
			name: "IconRepository.FindAllByUserIDs",
			run: func(t *testing.T, tx repository.Querier) ([]int64, []int64) {
				alice, bob := insertTestUser(t, tx, "alice"), insertTestUser(t, tx, "bob")
				for _, userID := range []domain.UserID{alice, bob} {
					if _, err := NewIconRepository().Create(ctx, tx, userID, []byte("icon")); err != nil {
						t.Fatalf("Create returned error: %v", err)
					}
				}
				icons, err := NewIconRepository().FindAllByUserIDs(ctx, tx, withManyIDs(alice, bob))
				if err != nil {
					t.Fatalf("FindAllByUserIDs returned error: %v", err)
				}
				got := make([]int64, len(icons))
				for i, icon := range icons {
					got[i] = int64(icon.UserID)
				}
				return int64s([]domain.UserID{alice, bob}), got
			},
		},
		{
			name: "LivestreamRepository.FindAllByIDs",
			run: func(t *testing.T, tx repository.Querier) ([]int64, []int64) {
				ownerID := insertTestUser(t, tx, "alice")
				first, second := insertTestLivestream(t, tx, ownerID, "first"), insertTestLivestream(t, tx, ownerID, "second")
				livestreams, err := NewLivestreamRepository().FindAllByIDs(ctx, tx, withManyIDs(first, second))
				if err != nil {
					t.Fatalf("FindAllByIDs returned error: %v", err)
				}
				got := make([]int64, len(livestreams))
				for i, l := range livestreams {
					got[i] = int64(l.ID)
				}
				return int64s([]domain.LivestreamID{first, second}), got
			},
		},
		{
			name: "LivestreamTagRepository.FindAllByLivestreamIDs",
			run: func(t *testing.T, tx repository.Querier) ([]int64, []int64) {
				ownerID := insertTestUser(t, tx, "alice")
				first, second := insertTestLivestream(t, tx, ownerID, "first"), insertTestLivestream(t, tx, ownerID, "second")
				insertTestTag(t, tx, first, "A")
				insertTestTag(t, tx, second, "B")
				livestreamTags, err := NewLivestreamTagRepository().FindAllByLivestreamIDs(ctx, tx, withManyIDs(first, second))
				if err != nil {
					t.Fatalf("FindAllByLivestreamIDs returned error: %v", err)
				}
				got := make([]int64, len(livestreamTags))
				for i, lt := range livestreamTags {
					got[i] = int64(lt.LivestreamID)
				}
				return int64s([]domain.LivestreamID{first, second}), got
			},
		},
		{
			name: "TagRepository.FindAllByIDs",
			run: func(t *testing.T, tx repository.Querier) ([]int64, []int64) {
				ownerID := insertTestUser(t, tx, "alice")
				livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
				a, b := insertTestTag(t, tx, livestreamID, "A"), insertTestTag(t, tx, livestreamID, "B")
				tags, err := NewTagRepository().FindAllByIDs(ctx, tx, withManyIDs(a.ID, b.ID))
				if err != nil {
					t.Fatalf("FindAllByIDs returned error: %v", err)
				}
				got := make([]int64, len(tags))
				for i, tag := range tags {
					got[i] = int64(tag.ID)
				}
				return int64s([]domain.TagID{a.ID, b.ID}), got
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := beginTestTx(t)
			want, got := tt.run(t, tx)
			slices.Sort(want)
			slices.Sort(got)
			if !slices.Equal(got, want) {
				t.Errorf("returned IDs = %v, want %v", got, want)
			}
		})
	}
}

// GORM に移行した repository が、移行前と同じカラムを読む SQL を発行していること (SELECT * にしない)。
func TestMigratedRepositories_IssueExplicitColumnSQL(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	sqls := recordSQL(t, tx)

	userID := insertTestUser(t, tx, "alice")
	tags := NewTagRepository()
	themes := NewThemeRepository()
	icons := NewIconRepository()
	users := NewUserRepository()
	livestreams := NewLivestreamRepository()
	livestreamTags := NewLivestreamTagRepository()
	slots := NewReservationSlotRepository()
	viewers := NewLivestreamViewersHistoryRepository()
	reports := NewLivecommentReportRepository()
	reactions := NewReactionRepository()
	ngWords := NewNGWordRepository()
	period := domain.ReservationPeriod{StartAt: 1, EndAt: 2}
	// 記録するのは、ここから後に発行された SQL
	*sqls = nil

	calls := []struct {
		name string
		call func() error
		want string
	}{
		{"tag FindAll", func() error { _, err := tags.FindAll(ctx, tx); return err },
			"SELECT id, name FROM `tags`"},
		{"tag FindIDsByName", func() error { _, err := tags.FindIDsByName(ctx, tx, "x"); return err },
			"SELECT `id` FROM `tags` WHERE name = ?"},
		{"tag FindAllByIDs", func() error { _, err := tags.FindAllByIDs(ctx, tx, []domain.TagID{1, 2}); return err },
			"SELECT id, name FROM `tags` WHERE id IN (?,?)"},
		{"theme FindByUserID", func() error { _, err := themes.FindByUserID(ctx, tx, userID); return err },
			"SELECT id, user_id, dark_mode FROM `themes` WHERE user_id = ? LIMIT ?"},
		{"theme FindAllByUserIDs", func() error { _, err := themes.FindAllByUserIDs(ctx, tx, []domain.UserID{userID}); return err },
			"SELECT id, user_id, dark_mode FROM `themes` WHERE user_id IN (?)"},
		{"theme Create", func() error { return themes.Create(ctx, tx, &domain.Theme{UserID: userID, DarkMode: true}) },
			"INSERT INTO `themes` (`user_id`,`dark_mode`) VALUES (?,?)"},
		{"icon FindImageByUserID", func() error { _, err := icons.FindImageByUserID(ctx, tx, userID); return err },
			"SELECT `image` FROM `icons` WHERE user_id = ? ORDER BY `icons`.`id` LIMIT ?"},
		{"icon FindAllByUserIDs", func() error {
			_, err := icons.FindAllByUserIDs(ctx, tx, []domain.UserID{userID, userID + 1})
			return err
		},
			"SELECT id, user_id, image FROM `icons` WHERE user_id IN (?,?) ORDER BY id"},
		{"icon Create", func() error { _, err := icons.Create(ctx, tx, userID, []byte("img")); return err },
			"INSERT INTO `icons` (`user_id`,`image`) VALUES (?,?)"},
		{"icon DeleteByUserID", func() error { return icons.DeleteByUserID(ctx, tx, userID) },
			"DELETE FROM `icons` WHERE user_id = ?"},
		{"user FindIDByName", func() error { _, err := users.FindIDByName(ctx, tx, "nobody"); return err },
			"SELECT `id` FROM `users` WHERE name = ? LIMIT ?"},
		{"user FindByName", func() error { _, err := users.FindByName(ctx, tx, "nobody"); return err },
			"SELECT id, name, display_name, description, password FROM `users` WHERE name = ? LIMIT ?"},
		{"user FindByID", func() error { _, err := users.FindByID(ctx, tx, userID); return err },
			"SELECT id, name, display_name, description, password FROM `users` WHERE id = ? LIMIT ?"},
		{"user FindAll", func() error { _, err := users.FindAll(ctx, tx); return err },
			"SELECT id, name, display_name, description, password FROM `users`"},
		{"user FindAllByIDs", func() error { _, err := users.FindAllByIDs(ctx, tx, []domain.UserID{userID, userID + 1}); return err },
			"SELECT id, name, display_name, description, password FROM `users` WHERE id IN (?,?)"},
		{"livestream FindByID", func() error { _, err := livestreams.FindByID(ctx, tx, 999999); return err },
			"SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM `livestreams` WHERE id = ? LIMIT ?"},
		{"livestream FindAllByIDAndUserID", func() error { _, err := livestreams.FindAllByIDAndUserID(ctx, tx, 1, userID); return err },
			"SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM `livestreams` WHERE id = ? AND user_id = ?"},
		{"livestream FindAll", func() error { _, err := livestreams.FindAll(ctx, tx); return err },
			"SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM `livestreams`"},
		{"livestream FindAllByUserID", func() error { _, err := livestreams.FindAllByUserID(ctx, tx, userID); return err },
			"SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM `livestreams` WHERE user_id = ?"},
		{"livestream FindAllOrderByIDDesc", func() error { _, err := livestreams.FindAllOrderByIDDesc(ctx, tx); return err },
			"SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM `livestreams` ORDER BY id DESC"},
		{"livestream FindAllOrderByIDDescLimited", func() error { _, err := livestreams.FindAllOrderByIDDescLimited(ctx, tx, 5); return err },
			"SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM `livestreams` ORDER BY id DESC LIMIT ?"},
		{"livestream FindAllByIDs", func() error { _, err := livestreams.FindAllByIDs(ctx, tx, []domain.LivestreamID{1, 2}); return err },
			"SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM `livestreams` WHERE id IN (?,?)"},
		{"livestream Create", func() error {
			_, err := livestreams.Create(ctx, tx, &domain.Livestream{UserID: userID, Title: "t", Description: "d", PlaylistUrl: "p", ThumbnailUrl: "u", StartAt: 1, EndAt: 2})
			return err
		}, "INSERT INTO `livestreams` (`user_id`,`title`,`description`,`playlist_url`,`thumbnail_url`,`start_at`,`end_at`) VALUES (?,?,?,?,?,?,?)"},
		{"livestream tag Create", func() error { return livestreamTags.Create(ctx, tx, 1, 2) },
			"INSERT INTO `livestream_tags` (`livestream_id`,`tag_id`) VALUES (?,?)"},
		{"livestream tag FindAllByTagIDs", func() error { _, err := livestreamTags.FindAllByTagIDs(ctx, tx, []domain.TagID{1, 2}); return err },
			"SELECT id, livestream_id, tag_id FROM `livestream_tags` WHERE tag_id IN (?,?) ORDER BY livestream_id DESC"},
		{"livestream tag FindAllByLivestreamIDs", func() error {
			_, err := livestreamTags.FindAllByLivestreamIDs(ctx, tx, []domain.LivestreamID{1, 2})
			return err
		}, "SELECT id, livestream_id, tag_id FROM `livestream_tags` WHERE livestream_id IN (?,?) ORDER BY id"},
		{"reservation slot FindAllByRangeForUpdate", func() error { _, err := slots.FindAllByRangeForUpdate(ctx, tx, period); return err },
			"SELECT id, slot, start_at, end_at FROM `reservation_slots` WHERE start_at >= ? AND end_at <= ? FOR UPDATE"},
		{"reservation slot FindSlotByStartAtAndEndAt", func() error { _, err := slots.FindSlotByStartAtAndEndAt(ctx, tx, 1, 2); return err },
			"SELECT `slot` FROM `reservation_slots` WHERE start_at = ? AND end_at = ? LIMIT ?"},
		{"reservation slot DecrementSlotsByRange", func() error { return slots.DecrementSlotsByRange(ctx, tx, period) },
			"UPDATE `reservation_slots` SET `slot`=slot - 1 WHERE start_at >= ? AND end_at <= ?"},
		{"viewers history Create", func() error {
			return viewers.Create(ctx, tx, &domain.LivestreamViewersHistory{UserID: userID, LivestreamID: 1, CreatedAt: 100})
		}, "INSERT INTO `livestream_viewers_history` (`user_id`,`livestream_id`,`created_at`) VALUES (?,?,?)"},
		{"viewers history DeleteByUserIDAndLivestreamID", func() error { return viewers.DeleteByUserIDAndLivestreamID(ctx, tx, userID, 1) },
			"DELETE FROM `livestream_viewers_history` WHERE user_id = ? AND livestream_id = ?"},
		{"viewers history CountByLivestreamID", func() error { _, err := viewers.CountByLivestreamID(ctx, tx, 1); return err },
			"SELECT count(*) FROM `livestream_viewers_history` WHERE livestream_id = ?"},
		{"viewers history CountViewersByLivestreamID", func() error { _, err := viewers.CountViewersByLivestreamID(ctx, tx, 1); return err },
			"SELECT count(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?"},
		{"report FindByID", func() error { _, err := reports.FindByID(ctx, tx, 999999); return err },
			"SELECT id, user_id, livestream_id, livecomment_id, created_at FROM `livecomment_reports` WHERE id = ? LIMIT ?"},
		{"report FindAllByLivestreamID", func() error { _, err := reports.FindAllByLivestreamID(ctx, tx, 1); return err },
			"SELECT id, user_id, livestream_id, livecomment_id, created_at FROM `livecomment_reports` WHERE livestream_id = ?"},
		{"report Create", func() error {
			_, err := reports.Create(ctx, tx, &domain.LivecommentReport{UserID: userID, LivestreamID: 1, LivecommentID: 2, CreatedAt: 100})
			return err
		}, "INSERT INTO `livecomment_reports` (`user_id`,`livestream_id`,`livecomment_id`,`created_at`) VALUES (?,?,?,?)"},
		{"report CountByLivestreamID", func() error { _, err := reports.CountByLivestreamID(ctx, tx, 1); return err },
			"SELECT count(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?"},
		{"reaction FindByID", func() error { _, err := reactions.FindByID(ctx, tx, 999999); return err },
			"SELECT id, emoji_name, user_id, livestream_id, created_at FROM `reactions` WHERE id = ? LIMIT ?"},
		{"reaction FindAllByLivestreamIDOrderByCreatedAtDesc", func() error {
			_, err := reactions.FindAllByLivestreamIDOrderByCreatedAtDesc(ctx, tx, 1)
			return err
		}, "SELECT id, emoji_name, user_id, livestream_id, created_at FROM `reactions` WHERE livestream_id = ? ORDER BY created_at DESC"},
		{"reaction FindAllByLivestreamIDOrderByCreatedAtDescLimited", func() error {
			_, err := reactions.FindAllByLivestreamIDOrderByCreatedAtDescLimited(ctx, tx, 1, 5)
			return err
		}, "SELECT id, emoji_name, user_id, livestream_id, created_at FROM `reactions` WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?"},
		{"reaction Create", func() error {
			_, err := reactions.Create(ctx, tx, &domain.Reaction{UserID: userID, LivestreamID: 1, EmojiName: "tada", CreatedAt: 100})
			return err
		}, "INSERT INTO `reactions` (`user_id`,`livestream_id`,`emoji_name`,`created_at`) VALUES (?,?,?,?)"},
		{"reaction CountByLivestreamOwnerID", func() error { _, err := reactions.CountByLivestreamOwnerID(ctx, tx, userID); return err },
			"SELECT count(*) FROM users u INNER JOIN livestreams l ON l.user_id = u.id INNER JOIN reactions r ON r.livestream_id = l.id WHERE u.id = ?"},
		{"reaction CountByLivestreamOwnerName", func() error { _, err := reactions.CountByLivestreamOwnerName(ctx, tx, "alice"); return err },
			"SELECT count(*) FROM users u INNER JOIN livestreams l ON l.user_id = u.id INNER JOIN reactions r ON r.livestream_id = l.id WHERE u.name = ?"},
		{"reaction FindFavoriteEmojiByLivestreamOwnerName", func() error {
			_, err := reactions.FindFavoriteEmojiByLivestreamOwnerName(ctx, tx, "nobody")
			return err
		}, "SELECT r.emoji_name FROM users u INNER JOIN livestreams l ON l.user_id = u.id INNER JOIN reactions r ON r.livestream_id = l.id WHERE u.name = ? GROUP BY `emoji_name` ORDER BY COUNT(*) DESC, emoji_name DESC LIMIT ?"},
		{"reaction CountByLivestreamID", func() error { _, err := reactions.CountByLivestreamID(ctx, tx, 1); return err },
			"SELECT count(*) FROM livestreams l INNER JOIN reactions r ON l.id = r.livestream_id WHERE l.id = ?"},
		{"reaction CountTotalByLivestreamID", func() error { _, err := reactions.CountTotalByLivestreamID(ctx, tx, 1); return err },
			"SELECT count(*) FROM livestreams l INNER JOIN reactions r ON r.livestream_id = l.id WHERE l.id = ?"},
		{"ng word FindAllByLivestreamID", func() error { _, err := ngWords.FindAllByLivestreamID(ctx, tx, 1); return err },
			"SELECT id, user_id, livestream_id, word, created_at FROM `ng_words` WHERE livestream_id = ?"},
		{"ng word FindAllByUserIDAndLivestreamID", func() error {
			_, err := ngWords.FindAllByUserIDAndLivestreamID(ctx, tx, userID, 1)
			return err
		}, "SELECT id, user_id, livestream_id, word, created_at FROM `ng_words` WHERE user_id = ? AND livestream_id = ? ORDER BY created_at DESC"},
		{"ng word Matches (raw)", func() error { _, err := ngWords.Matches(ctx, tx, "spam", "sp"); return err },
			"SELECT COUNT(*) FROM (SELECT ? AS text) AS texts INNER JOIN (SELECT CONCAT('%', ?, '%') AS pattern) AS patterns ON texts.text LIKE patterns.pattern;"},
		{"ng word Create", func() error {
			_, err := ngWords.Create(ctx, tx, &domain.NGWord{UserID: userID, LivestreamID: 1, Word: "w", CreatedAt: 100})
			return err
		}, "INSERT INTO `ng_words` (`user_id`,`livestream_id`,`word`,`created_at`) VALUES (?,?,?,?)"},
		{"user Create", func() error {
			_, err := users.Create(ctx, tx, &domain.User{Name: "carol-sql", DisplayName: "Carol", Description: "d", HashedPassword: "hashed"})
			return err
		}, "INSERT INTO `users` (`name`,`display_name`,`description`,`password`) VALUES (?,?,?,?)"},
	}
	for _, c := range calls {
		*sqls = nil
		err := c.call()
		// theme / icon / user / livestream の FindXxx は、無いものを引くので ErrNotFound (予約枠は sql.ErrNoRows) になるが、SQL は発行されている
		if err != nil && !errors.Is(err, repository.ErrNotFound) && !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("%s returned error: %v", c.name, err)
		}
		// Raw のリテラルは改行やタブを含むので、空白の違いは無視して比べる
		if len(*sqls) != 1 || strings.Join(strings.Fields((*sqls)[0]), " ") != c.want {
			t.Errorf("%s: SQL = %q, want [%q]", c.name, *sqls, c.want)
		}
	}
}

// 分割して引いても、紐付けは結果全体で ID の昇順になること。
func TestLivestreamTagRepository_FindAllByLivestreamIDs_KeepsOrderAcrossChunks(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	var livestreamIDs []domain.LivestreamID
	// 分割する大きさ (maxInArgs) を超えるライブ配信にタグを付ける。ID が大きいライブ配信ほど先に紐付ける
	for i := 0; i < maxInArgs+5; i++ {
		livestreamIDs = append(livestreamIDs, insertTestLivestream(t, tx, ownerID, "stream"))
	}
	for i := len(livestreamIDs) - 1; i >= 0; i-- {
		insertTestTag(t, tx, livestreamIDs[i], fmt.Sprintf("tag-%d", i))
	}

	got, err := NewLivestreamTagRepository().FindAllByLivestreamIDs(ctx, tx, livestreamIDs)
	if err != nil {
		t.Fatalf("FindAllByLivestreamIDs returned error: %v", err)
	}
	if len(got) != len(livestreamIDs) {
		t.Fatalf("len(livestream tags) = %d, want %d", len(got), len(livestreamIDs))
	}
	if !slices.IsSortedFunc(got, func(a, b *domain.LivestreamTag) int { return int(a.ID) - int(b.ID) }) {
		t.Errorf("livestream tags are not sorted by ID")
	}
}

// 空の tagIDs は、移行前と同じくエラーになること (GORM の IN ? は、空だとエラーにならずに何も返さないため、明示的に弾いている)。
func TestLivestreamTagRepository_FindAllByTagIDs_EmptyIsError(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewLivestreamTagRepository().FindAllByTagIDs(context.Background(), tx, nil)
	if err == nil || err.Error() != "failed to construct IN query: empty slice passed to 'in' query" {
		t.Errorf("err = %v, want the empty IN error", err)
	}
}

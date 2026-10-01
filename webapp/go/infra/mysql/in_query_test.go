package mysql

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestSelectIn_Empty(t *testing.T) {
	tx := beginTestTx(t)

	got, err := selectIn[domain.User](context.Background(), tx, "SELECT id, name, display_name, description, password FROM users WHERE id IN (?)", []domain.UserID(nil))
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("selectIn(nil) = %#v, %v, want empty non-nil slice", got, err)
	}
}

func TestSelectInChunked_ConcatenatesChunks(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	var ids []domain.UserID
	for _, name := range []string{"alice", "bob", "carol", "dave", "eve"} {
		ids = append(ids, insertTestUser(t, tx, name))
	}

	// 2 件ずつに分けて引いても、全ての行が 1 回ずつ返る
	got, err := selectInChunked[domain.User](ctx, tx, "SELECT id, name, display_name, description, password FROM users WHERE id IN (?)", ids, 2)
	if err != nil {
		t.Fatalf("selectInChunked returned error: %v", err)
	}
	gotIDs := make([]domain.UserID, len(got))
	for i, u := range got {
		gotIDs[i] = u.ID
	}
	slices.Sort(gotIDs)
	if !slices.Equal(gotIDs, ids) {
		t.Errorf("IDs = %v, want %v", gotIDs, ids)
	}
}

// MySQL のプレースホルダーの上限 (65535) を超える数の ID でも、分割して引けること。
func TestSelectIn_MoreIDsThanMySQLPlaceholderLimit(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	alice := insertTestUser(t, tx, "alice")
	bob := insertTestUser(t, tx, "bob")
	ids := []domain.UserID{alice}
	// 存在しない ID で水増しする
	for id := domain.UserID(1_000_000); len(ids) < 70_000; id++ {
		ids = append(ids, id)
	}
	ids = append(ids, bob)

	got, err := NewUserRepository().FindAllByIDs(ctx, tx, ids)
	if err != nil {
		t.Fatalf("FindAllByIDs returned error: %v", err)
	}
	names := map[string]bool{}
	for _, u := range got {
		names[u.Name] = true
	}
	if len(got) != 2 || !names["alice"] || !names["bob"] {
		t.Errorf("users = %+v, want alice and bob", got)
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

// ID が順不同・重複ありでも、全ての行が 1 回ずつ返り、呼び出し側の ID の一覧は変わらないこと。
func TestSelectIn_UnsortedAndDuplicateIDs(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	alice := insertTestUser(t, tx, "alice")
	bob := insertTestUser(t, tx, "bob")
	carol := insertTestUser(t, tx, "carol")
	ids := []domain.UserID{carol, alice, carol, bob, alice}
	original := slices.Clone(ids)

	got, err := selectIn[domain.User](ctx, tx, "SELECT id, name, display_name, description, password FROM users WHERE id IN (?)", ids)
	if err != nil {
		t.Fatalf("selectIn returned error: %v", err)
	}
	gotIDs := make([]domain.UserID, len(got))
	for i, u := range got {
		gotIDs[i] = u.ID
	}
	// 重複は除かれ、ID の昇順に引く (ORDER BY が無くても主キーの順に返る)
	if want := []domain.UserID{alice, bob, carol}; !slices.Equal(gotIDs, want) {
		t.Errorf("IDs = %v, want %v", gotIDs, want)
	}
	if !slices.Equal(ids, original) {
		t.Errorf("ids = %v, want unchanged %v", ids, original)
	}
}

// recordingQuerier は SelectContext に渡された引数を記録する repository.Querier。結果の行は返さない。
type recordingQuerier struct {
	repository.Querier
	// params はクエリごとの引数 (IN に展開された ID)。
	params [][]any
}

func (r *recordingQuerier) SelectContext(_ context.Context, _ any, _ string, args ...any) error {
	r.params = append(r.params, args)
	return nil
}

// DB に渡る ID は重複を除いて昇順に並び、その順に分割されること。
func TestSelectInChunked_SendsSortedDistinctIDsInChunks(t *testing.T) {
	q := &recordingQuerier{}

	_, err := selectInChunked[domain.User](context.Background(), q, "SELECT id FROM users WHERE id IN (?)", []domain.UserID{3, 1, 3, 5, 2, 1, 4}, 2)
	if err != nil {
		t.Fatalf("selectInChunked returned error: %v", err)
	}

	want := [][]any{{domain.UserID(1), domain.UserID(2)}, {domain.UserID(3), domain.UserID(4)}, {domain.UserID(5)}}
	if !reflect.DeepEqual(q.params, want) {
		t.Errorf("params = %v, want %v", q.params, want)
	}
}

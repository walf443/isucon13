package mysql

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	driver "github.com/go-sql-driver/mysql"
	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func insertTestReservationSlot(t *testing.T, tx repository.Querier, slot int64, startAt int64, endAt int64) domain.ReservationSlotID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(), "INSERT INTO reservation_slots (slot, start_at, end_at) VALUES (?, ?, ?)", slot, startAt, endAt)
	if err != nil {
		t.Fatalf("failed to insert reservation slot: %v", err)
	}
	id, _ := res.LastInsertId()
	return domain.ReservationSlotID(id)
}

// 他のテストのデータと重ならないよう、遠い将来の時刻を使う
const (
	testSlotBase = 5000000000
	testSlotHour = 3600
)

func TestReservationSlotRepository_FindAllByRangeForUpdate(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	before := insertTestReservationSlot(t, tx, 5, testSlotBase-testSlotHour, testSlotBase)
	first := insertTestReservationSlot(t, tx, 5, testSlotBase, testSlotBase+testSlotHour)
	second := insertTestReservationSlot(t, tx, 3, testSlotBase+testSlotHour, testSlotBase+2*testSlotHour)
	after := insertTestReservationSlot(t, tx, 5, testSlotBase+2*testSlotHour, testSlotBase+3*testSlotHour)

	slots, err := NewReservationSlotRepository().FindAllByRangeForUpdate(ctx, tx, domain.ReservationPeriod{StartAt: testSlotBase, EndAt: testSlotBase + 2*testSlotHour})
	if err != nil {
		t.Fatalf("FindAllByRangeForUpdate returned error: %v", err)
	}
	got := map[domain.ReservationSlotID]domain.ReservationSlot{}
	for _, s := range slots {
		got[s.ID] = *s
	}
	if len(got) != 2 {
		t.Fatalf("slots = %+v, want 2 slots (excluding %d, %d)", got, before, after)
	}
	if want := (domain.ReservationSlot{ID: first, Slot: 5, StartAt: testSlotBase, EndAt: testSlotBase + testSlotHour}); got[first] != want {
		t.Errorf("slot = %+v, want %+v", got[first], want)
	}
	if want := (domain.ReservationSlot{ID: second, Slot: 3, StartAt: testSlotBase + testSlotHour, EndAt: testSlotBase + 2*testSlotHour}); got[second] != want {
		t.Errorf("slot = %+v, want %+v", got[second], want)
	}
}

func TestReservationSlotRepository_FindSlotByStartAtAndEndAt(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewReservationSlotRepository()

	insertTestReservationSlot(t, tx, 4, testSlotBase, testSlotBase+testSlotHour)

	count, err := repo.FindSlotByStartAtAndEndAt(ctx, tx, testSlotBase, testSlotBase+testSlotHour)
	if err != nil {
		t.Fatalf("FindSlotByStartAtAndEndAt returned error: %v", err)
	}
	if count != 4 {
		t.Errorf("count = %d, want 4", count)
	}

	// 該当が無い場合は ErrNotFound に変換せず sql.ErrNoRows のまま返す (移行前と同じく 500 になる)
	if _, err := repo.FindSlotByStartAtAndEndAt(ctx, tx, testSlotBase, testSlotBase+2*testSlotHour); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestReservationSlotRepository_DecrementSlotsByRange(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	before := insertTestReservationSlot(t, tx, 5, testSlotBase-testSlotHour, testSlotBase)
	first := insertTestReservationSlot(t, tx, 5, testSlotBase, testSlotBase+testSlotHour)
	second := insertTestReservationSlot(t, tx, 3, testSlotBase+testSlotHour, testSlotBase+2*testSlotHour)
	after := insertTestReservationSlot(t, tx, 5, testSlotBase+2*testSlotHour, testSlotBase+3*testSlotHour)

	if err := NewReservationSlotRepository().DecrementSlotsByRange(ctx, tx, domain.ReservationPeriod{StartAt: testSlotBase, EndAt: testSlotBase + 2*testSlotHour}); err != nil {
		t.Fatalf("DecrementSlotsByRange returned error: %v", err)
	}

	for id, want := range map[domain.ReservationSlotID]int64{before: 5, first: 4, second: 2, after: 5} {
		var slot int64
		if err := tx.GetContext(ctx, &slot, "SELECT slot FROM reservation_slots WHERE id = ?", id); err != nil {
			t.Fatalf("failed to get reservation slot: %v", err)
		}
		if slot != want {
			t.Errorf("slot of %d = %d, want %d", id, slot, want)
		}
	}
}

// FindAllByRangeForUpdate が、範囲に収まる予約枠の行を本当にロックすること。
// 別のトランザクションから見えるように、コミット済みのデータを使い、別のトランザクションが NOWAIT でロックを取ろうとして確かめる。
func TestReservationSlotRepository_FindAllByRangeForUpdate_LocksRows(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	// 他のテストのデータと重ならない時刻を使う。コミットするので、終了時に必ず削除する
	const lockBase = 7000000000
	if err := testDB.Exec("INSERT INTO reservation_slots (slot, start_at, end_at) VALUES (?, ?, ?)", 5, lockBase, lockBase+testSlotHour).Error; err != nil {
		t.Fatalf("failed to insert reservation slot: %v", err)
	}
	other := testDB.Begin()
	t.Cleanup(func() {
		// ロックを持つトランザクションを先に終わらせてから、コミット済みのデータを消す (逆だと、ロックが外れるまで削除が待たされる)
		_ = gormOf(tx).Rollback().Error
		_ = other.Rollback().Error
		_ = testDB.Exec("DELETE FROM reservation_slots WHERE start_at = ?", lockBase).Error
	})
	lockNowait := func() error {
		var ids []int64
		return other.Raw("SELECT id FROM reservation_slots WHERE start_at = ? FOR UPDATE NOWAIT", lockBase).Scan(&ids).Error
	}

	if _, err := NewReservationSlotRepository().FindAllByRangeForUpdate(ctx, tx, domain.ReservationPeriod{StartAt: lockBase, EndAt: lockBase + testSlotHour}); err != nil {
		t.Fatalf("FindAllByRangeForUpdate returned error: %v", err)
	}

	// ロックされているので、別のトランザクションはすぐに失敗する (3572: NOWAIT でロックを取れなかった)
	if mysqlErr, ok := errors.AsType[*driver.MySQLError](lockNowait()); !ok || mysqlErr.Number != 3572 {
		t.Errorf("while locked: err is not a MySQL error with number 3572")
	}
	// ロックを持つトランザクションが終われば、取れる (さっき失敗したのが、このトランザクションのロックのせいだったということ)
	if err := gormOf(tx).Rollback().Error; err != nil {
		t.Fatalf("failed to roll back: %v", err)
	}
	if err := lockNowait(); err != nil {
		t.Errorf("after rollback: err = %v, want nil", err)
	}
}

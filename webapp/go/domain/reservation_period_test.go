package domain

import (
	"testing"
	"time"
)

func TestReservableTerm(t *testing.T) {
	// 移行前は time.Date で定義していたので、同じ時刻であることを確認する
	if want := time.Date(2023, 11, 25, 1, 0, 0, 0, time.UTC).Unix(); ReservableTerm.StartAt != want {
		t.Errorf("StartAt = %d, want %d", ReservableTerm.StartAt, want)
	}
	if want := time.Date(2024, 11, 25, 1, 0, 0, 0, time.UTC).Unix(); ReservableTerm.EndAt != want {
		t.Errorf("EndAt = %d, want %d", ReservableTerm.EndAt, want)
	}
}

func TestReservationPeriod_IsReservable(t *testing.T) {
	// 予約可能期間は 1700874000 〜 1732496400
	tests := []struct {
		name   string
		period ReservationPeriod
		want   bool
	}{
		{name: "inside the term", period: ReservationPeriod{StartAt: 1700874000, EndAt: 1700881200}, want: true},
		{name: "ends at the term start", period: ReservationPeriod{StartAt: 1700870400, EndAt: 1700874000}, want: false},
		{name: "ends before the term start", period: ReservationPeriod{StartAt: 1700866800, EndAt: 1700870400}, want: false},
		{name: "starts at the term end", period: ReservationPeriod{StartAt: 1732496400, EndAt: 1732500000}, want: false},
		{name: "starts after the term end", period: ReservationPeriod{StartAt: 1732500000, EndAt: 1732503600}, want: false},
		// 期間に一部でも掛かっていれば予約できる (移行前と同じ)
		{name: "overlaps the term start", period: ReservationPeriod{StartAt: 1700870400, EndAt: 1700877600}, want: true},
		{name: "overlaps the term end", period: ReservationPeriod{StartAt: 1732492800, EndAt: 1732500000}, want: true},
		{name: "covers the whole term", period: ReservationPeriod{StartAt: 1700000000, EndAt: 1800000000}, want: true},
		// 開始と終了が逆でも検証しない (移行前と同じく「開始 < 期間の終了 かつ 終了 > 期間の開始」だけで決まる)
		{name: "reversed inside the term", period: ReservationPeriod{StartAt: 1700881200, EndAt: 1700877600}, want: true},
		{name: "reversed ending at the term start", period: ReservationPeriod{StartAt: 1700881200, EndAt: 1700874000}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.period.IsReservable(); got != tt.want {
				t.Errorf("IsReservable() = %v, want %v", got, tt.want)
			}
		})
	}
}

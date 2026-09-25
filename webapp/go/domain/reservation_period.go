package domain

// ReservationPeriod は配信予約の区間 (開始時刻・終了時刻の UNIX 秒)。
//
// 移行前と同じく StartAt < EndAt かどうかは検証しない (開始と終了が逆の区間もそのまま扱う)。
type ReservationPeriod struct {
	StartAt int64
	EndAt   int64
}

// ReservableTerm は予約を受け付ける期間 (2023/11/25 10:00 JST からの1年間)。
var ReservableTerm = ReservationPeriod{
	StartAt: 1700874000, // 2023-11-25T01:00:00Z
	EndAt:   1732496400, // 2024-11-25T01:00:00Z
}

// Overlaps は p と other の区間が一部でも重なるかを返す。
// 区間の端がちょうど接しているだけの場合は重ならないとみなす。
func (p ReservationPeriod) Overlaps(other ReservationPeriod) bool {
	return p.StartAt < other.EndAt && other.StartAt < p.EndAt
}

// IsReservable は予約を受け付ける期間 (ReservableTerm) に一部でも掛かっているかを返す。
func (p ReservationPeriod) IsReservable() bool {
	return p.Overlaps(ReservableTerm)
}

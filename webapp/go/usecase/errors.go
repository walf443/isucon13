package usecase

import (
	"errors"
	"fmt"
)

// ErrUserNotFound は指定されたユーザが存在しないことを表す。
var ErrUserNotFound = errors.New("user not found")

// ErrReservedUsername は予約済みのユーザ名で登録しようとしたことを表す。
var ErrReservedUsername = errors.New("the username is reserved")

// ErrInvalidCredentials はユーザ名かパスワードが間違っていることを表す。
var ErrInvalidCredentials = errors.New("invalid username or password")

// ErrIconNotFound はユーザのアイコンが登録されていないことを表す。
var ErrIconNotFound = errors.New("icon not found")

// ErrLivestreamNotFound は指定されたライブ配信が存在しないことを表す。
var ErrLivestreamNotFound = errors.New("livestream not found")

// ErrBadReservationTimeRange は予約区間が予約可能期間に掛かっていないことを表す。
var ErrBadReservationTimeRange = errors.New("bad reservation time range")

// ReservationSlotUnavailableError は予約区間 StartAt 〜 EndAt に空きの無い予約枠があることを表す。
type ReservationSlotUnavailableError struct {
	StartAt int64
	EndAt   int64
}

func (e *ReservationSlotUnavailableError) Error() string {
	return fmt.Sprintf("予約期間 %d ~ %dに対して、予約区間 %d ~ %dが予約できません", reservationTermStartAt.Unix(), reservationTermEndAt.Unix(), e.StartAt, e.EndAt)
}

// ErrLivecommentNotFound は指定されたライブコメントが存在しないことを表す。
var ErrLivecommentNotFound = errors.New("livecomment not found")

// ErrNotLivestreamOwner はライブ配信の配信者でないユーザが、配信者向けの操作をしようとしたことを表す。
var ErrNotLivestreamOwner = errors.New("not the owner of the livestream")

// ErrSpamLivecomment はライブコメントが配信者の NG ワードに当たったことを表す。
var ErrSpamLivecomment = errors.New("livecomment is judged as spam")

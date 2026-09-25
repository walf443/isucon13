package usecase

import "errors"

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

// ErrBadReservationTimeRange は予約区間が予約可能期間 (ReservationTermStartAt 〜 ReservationTermEndAt) に掛かっていないことを表す。
var ErrBadReservationTimeRange = errors.New("bad reservation time range")

// ErrReservationSlotUnavailable は予約区間に空きの無い予約枠があることを表す。
var ErrReservationSlotUnavailable = errors.New("reservation slot is unavailable")

// ErrLivecommentNotFound は指定されたライブコメントが存在しないことを表す。
var ErrLivecommentNotFound = errors.New("livecomment not found")

// ErrNotLivestreamOwner はライブ配信の配信者でないユーザが、配信者向けの操作をしようとしたことを表す。
var ErrNotLivestreamOwner = errors.New("not the owner of the livestream")

// ErrSpamLivecomment はライブコメントが配信者の NG ワードに当たったことを表す。
var ErrSpamLivecomment = errors.New("livecomment is judged as spam")

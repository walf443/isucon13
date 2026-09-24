package usecase

import "errors"

// ErrUserNotFound は指定されたユーザが存在しないことを表す。
var ErrUserNotFound = errors.New("user not found")

// ErrIconNotFound はユーザのアイコンが登録されていないことを表す。
var ErrIconNotFound = errors.New("icon not found")

// ErrLivestreamNotFound は指定されたライブ配信が存在しないことを表す。
var ErrLivestreamNotFound = errors.New("livestream not found")

package usecase

import "errors"

// ErrUserNotFound は指定されたユーザが存在しないことを表す。
var ErrUserNotFound = errors.New("user not found")

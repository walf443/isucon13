package usecase

import "errors"

// ErrUserNotFound は指定されたユーザが存在しないことを表す。
var ErrUserNotFound = errors.New("user not found")

// ErrIconNotFound はユーザのアイコンが登録されていないことを表す。
var ErrIconNotFound = errors.New("icon not found")

// ErrLivestreamNotFound は指定されたライブ配信が存在しないことを表す。
var ErrLivestreamNotFound = errors.New("livestream not found")

// ErrNotLivestreamOwner はライブ配信の配信者でないユーザが、配信者向けの操作をしようとしたことを表す。
var ErrNotLivestreamOwner = errors.New("not the owner of the livestream")

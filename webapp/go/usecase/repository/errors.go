package repository

import "errors"

// ErrNotFound は対象のレコードが存在しないことを表す。
var ErrNotFound = errors.New("not found")

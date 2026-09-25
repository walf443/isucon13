package model

import (
	"crypto/sha256"
	"fmt"
)

type IconID = ID[IconModel]

type IconModel struct {
	ID     IconID `db:"id"`
	UserID UserID `db:"user_id"`
	Image  []byte `db:"image"`
}

// IconHash はアイコン画像のハッシュ (sha256 の16進文字列) を返す。
func IconHash(image []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(image))
}

// UserIconHash はユーザのアイコンのハッシュを返す。
// アイコンを登録していない (registered が false) 場合は、既定のアイコンのハッシュ defaultIconHash を返す。
// 空の画像を登録している場合は、登録済みとして空の画像のハッシュを返す (移行前と同じ)。
func UserIconHash(image []byte, registered bool, defaultIconHash string) string {
	if !registered {
		return defaultIconHash
	}
	return IconHash(image)
}

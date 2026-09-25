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

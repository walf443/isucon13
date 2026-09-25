package script

import (
	"os/exec"

	"github.com/isucon/isucon13/webapp/go/usecase"
)

type initializer struct {
	// scriptPath は初期化スクリプトのパス。
	scriptPath string
}

// NewInitializer は初期化スクリプト (sql/init.sh) を実行する Initializer を返す。
func NewInitializer(scriptPath string) usecase.Initializer {
	return &initializer{scriptPath: scriptPath}
}

// Initialize はスクリプトを実行し、標準出力と標準エラー出力をまとめて返す。
func (i *initializer) Initialize() ([]byte, error) {
	// 移行前と同じく context でキャンセルしない exec.Command を使う
	return exec.Command(i.scriptPath).CombinedOutput()
}

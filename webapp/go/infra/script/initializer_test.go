package script

import (
	"os"
	"path/filepath"
	"testing"
)

func writeScript(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "init.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}
	return path
}

func TestInitializer_Initialize(t *testing.T) {
	path := writeScript(t, "echo out\necho err 1>&2\n")

	out, err := NewInitializer(path).Initialize()
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
	// 標準出力と標準エラー出力をまとめて返す
	if got := string(out); got != "out\nerr\n" {
		t.Errorf("output = %q, want %q", got, "out\nerr\n")
	}
}

func TestInitializer_Initialize_Error(t *testing.T) {
	path := writeScript(t, "echo failed 1>&2\nexit 3\n")

	out, err := NewInitializer(path).Initialize()
	if err == nil || err.Error() != "exit status 3" {
		t.Errorf("err = %v, want exit status 3", err)
	}
	// 失敗した場合も出力を返す (ログに出すため)
	if got := string(out); got != "failed\n" {
		t.Errorf("output = %q, want %q", got, "failed\n")
	}
}

package powerdns

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installFakePdnsutil は引数をファイルに書き出して exitCode で終了する pdnsutil を PATH に置く。
func installFakePdnsutil(t *testing.T, output string, exitCode string) string {
	t.Helper()
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	script := "#!/bin/sh\necho \"$@\" > " + argsFile + "\nprintf '%s' '" + output + "'\nexit " + exitCode + "\n"
	if err := os.WriteFile(filepath.Join(dir, "pdnsutil"), []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake pdnsutil: %v", err)
	}
	t.Setenv("PATH", dir)
	return argsFile
}

func TestDNSRecordRegistrar_AddRecord(t *testing.T) {
	argsFile := installFakePdnsutil(t, "", "0")

	if err := NewDNSRecordRegistrar("192.0.2.1").AddRecord("alice"); err != nil {
		t.Fatalf("AddRecord returned error: %v", err)
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read args: %v", err)
	}
	if got, want := strings.TrimSpace(string(args)), "add-record u.isucon.dev alice A 0 192.0.2.1"; got != want {
		t.Errorf("args = %q, want %q", got, want)
	}
}

func TestDNSRecordRegistrar_AddRecord_Error(t *testing.T) {
	installFakePdnsutil(t, "Error: zone not found", "1")

	err := NewDNSRecordRegistrar("192.0.2.1").AddRecord("alice")
	// 移行前のレスポンスと同じく "<出力>: <エラー>" の形式
	if err == nil || err.Error() != "Error: zone not found: exit status 1" {
		t.Errorf("err = %v, want %q", err, "Error: zone not found: exit status 1")
	}
}

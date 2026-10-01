package mysql

import (
	"os"
	"testing"
)

func TestConfig_dsn(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "default values",
			cfg:  Config{Net: "tcp", Addr: "127.0.0.1:3306", User: "isucon", Passwd: "isucon", DBName: "isupipe", ParseTime: true},
			want: "isucon:isucon@tcp(127.0.0.1:3306)/isupipe?parseTime=true",
		},
		{
			name: "without parseTime",
			cfg:  Config{Net: "tcp", Addr: "db:3307", User: "u", Passwd: "p", DBName: "d"},
			want: "u:p@tcp(db:3307)/d",
		},
		{
			// パスワードの記号はドライバの FormatDSN がそのまま扱う
			name: "password with symbols",
			cfg:  Config{Net: "tcp", Addr: "127.0.0.1:3306", User: "isucon", Passwd: "p@ss:w/rd", DBName: "isupipe", ParseTime: true},
			want: "isucon:p@ss:w/rd@tcp(127.0.0.1:3306)/isupipe?parseTime=true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.dsn(); got != tt.want {
				t.Errorf("dsn() = %q, want %q", got, tt.want)
			}
		})
	}
}

// unsetDialConfigEnv は ISUCON13_MYSQL_DIALCONFIG_* をテスト中だけ未設定にする。
func unsetDialConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"ISUCON13_MYSQL_DIALCONFIG_NET",
		"ISUCON13_MYSQL_DIALCONFIG_ADDRESS",
		"ISUCON13_MYSQL_DIALCONFIG_PORT",
		"ISUCON13_MYSQL_DIALCONFIG_USER",
		"ISUCON13_MYSQL_DIALCONFIG_PASSWORD",
		"ISUCON13_MYSQL_DIALCONFIG_DATABASE",
		"ISUCON13_MYSQL_DIALCONFIG_PARSETIME",
	} {
		// t.Setenv でテスト終了時に元の値へ戻るようにしてから消す
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("failed to unset %s: %v", key, err)
		}
	}
}

func TestConfigFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr string
	}{
		{
			name: "uses default values when no environment variables are set",
			want: Config{Net: "tcp", Addr: "127.0.0.1:3306", User: "isucon", Passwd: "isucon", DBName: "isupipe", ParseTime: true},
		},
		{
			name: "overrides values with environment variables",
			env: map[string]string{
				"ISUCON13_MYSQL_DIALCONFIG_NET":       "unix",
				"ISUCON13_MYSQL_DIALCONFIG_ADDRESS":   "db",
				"ISUCON13_MYSQL_DIALCONFIG_PORT":      "3307",
				"ISUCON13_MYSQL_DIALCONFIG_USER":      "u",
				"ISUCON13_MYSQL_DIALCONFIG_PASSWORD":  "p",
				"ISUCON13_MYSQL_DIALCONFIG_DATABASE":  "d",
				"ISUCON13_MYSQL_DIALCONFIG_PARSETIME": "false",
			},
			want: Config{Net: "unix", Addr: "db:3307", User: "u", Passwd: "p", DBName: "d", ParseTime: false},
		},
		{
			name: "uses port 3306 when only address is set",
			env:  map[string]string{"ISUCON13_MYSQL_DIALCONFIG_ADDRESS": "db"},
			want: Config{Net: "tcp", Addr: "db:3306", User: "isucon", Passwd: "isucon", DBName: "isupipe", ParseTime: true},
		},
		{
			name: "ignores port when address is not set",
			env:  map[string]string{"ISUCON13_MYSQL_DIALCONFIG_PORT": "3307"},
			want: Config{Net: "tcp", Addr: "127.0.0.1:3306", User: "isucon", Passwd: "isucon", DBName: "isupipe", ParseTime: true},
		},
		{
			name:    "returns error when parseTime is not bool",
			env:     map[string]string{"ISUCON13_MYSQL_DIALCONFIG_PARSETIME": "yes"},
			wantErr: `failed to parse environment variable 'ISUCON13_MYSQL_DIALCONFIG_PARSETIME' as bool: strconv.ParseBool: parsing "yes": invalid syntax`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsetDialConfigEnv(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := ConfigFromEnv()
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ConfigFromEnv() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

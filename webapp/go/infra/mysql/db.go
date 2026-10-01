package mysql

import (
	"fmt"
	"net"
	"os"
	"strconv"

	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config は MySQL への接続設定。
type Config struct {
	Net       string
	Addr      string
	User      string
	Passwd    string
	DBName    string
	ParseTime bool
}

// ConfigFromEnv は環境変数 ISUCON13_MYSQL_DIALCONFIG_* から接続設定を読み取る。
// 環境変数が無い項目には既定値を使う。
func ConfigFromEnv() (Config, error) {
	const (
		networkTypeEnvKey = "ISUCON13_MYSQL_DIALCONFIG_NET"
		addrEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_ADDRESS"
		portEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_PORT"
		userEnvKey        = "ISUCON13_MYSQL_DIALCONFIG_USER"
		passwordEnvKey    = "ISUCON13_MYSQL_DIALCONFIG_PASSWORD"
		dbNameEnvKey      = "ISUCON13_MYSQL_DIALCONFIG_DATABASE"
		parseTimeEnvKey   = "ISUCON13_MYSQL_DIALCONFIG_PARSETIME"
	)

	conf := Config{}

	// 環境変数がセットされていなかった場合でも一旦動かせるように、デフォルト値を入れておく
	// この挙動を変更して、エラーを出すようにしてもいいかもしれない
	conf.Net = "tcp"
	conf.Addr = net.JoinHostPort("127.0.0.1", "3306")
	conf.User = "isucon"
	conf.Passwd = "isucon"
	conf.DBName = "isupipe"
	conf.ParseTime = true

	if v, ok := os.LookupEnv(networkTypeEnvKey); ok {
		conf.Net = v
	}
	if addr, ok := os.LookupEnv(addrEnvKey); ok {
		if port, ok2 := os.LookupEnv(portEnvKey); ok2 {
			conf.Addr = net.JoinHostPort(addr, port)
		} else {
			conf.Addr = net.JoinHostPort(addr, "3306")
		}
	}
	if v, ok := os.LookupEnv(userEnvKey); ok {
		conf.User = v
	}
	if v, ok := os.LookupEnv(passwordEnvKey); ok {
		conf.Passwd = v
	}
	if v, ok := os.LookupEnv(dbNameEnvKey); ok {
		conf.DBName = v
	}
	if v, ok := os.LookupEnv(parseTimeEnvKey); ok {
		parseTime, err := strconv.ParseBool(v)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse environment variable '%s' as bool: %+v", parseTimeEnvKey, err)
		}
		conf.ParseTime = parseTime
	}

	return conf, nil
}

// Open は cfg の MySQL に接続し、疎通を確認した *gorm.DB を返す。
// 疎通の確認 (ping) は gorm.Open が行う (DisableAutomaticPing を無効のままにしている) ので、ここでは重ねて行わない。
func Open(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(gormmysql.Open(cfg.dsn()), gormConfig())
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)

	return db, nil
}

// Close は Open した *gorm.DB の接続を閉じる。
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// dsn はドライバの既定値を土台に cfg の値で上書きした DSN を返す。
func (cfg Config) dsn() string {
	conf := driver.NewConfig()
	conf.Net = cfg.Net
	conf.Addr = cfg.Addr
	conf.User = cfg.User
	conf.Passwd = cfg.Passwd
	conf.DBName = cfg.DBName
	conf.ParseTime = cfg.ParseTime
	return conf.FormatDSN()
}

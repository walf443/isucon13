package main

// ISUCON的な参考: https://github.com/isucon/isucon12-qualify/blob/main/webapp/go/isuports.go#L336
// sqlx的な参考: https://jmoiron.github.io/sqlx/

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/isucon/isucon13/webapp/go/infra/mysql"
	"github.com/isucon/isucon13/webapp/go/interfaces/http/handler"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	echolog "github.com/labstack/gommon/log"
)

const (
	listenPort                     = 8080
	powerDNSSubdomainAddressEnvKey = "ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS"
)

var (
	secret        = []byte("isucon13_session_cookiestore_defaultsecret")
	fallbackImage = "../img/NoImage.jpg"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if secretKey, ok := os.LookupEnv("ISUCON13_SESSION_SECRETKEY"); ok {
		secret = []byte(secretKey)
	}
}

func connectDB(logger echo.Logger) (*sqlx.DB, error) {
	conf, err := mysql.ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return mysql.Open(conf)
}

func main() {
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(echolog.DEBUG)
	e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	// DB接続
	conn, err := connectDB(e.Logger)
	if err != nil {
		e.Logger.Errorf("failed to connect db: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	// ユーザ登録で使うので、handler を組み立てる前に読み込む
	subdomainAddr, ok := os.LookupEnv(powerDNSSubdomainAddressEnvKey)
	if !ok {
		e.Logger.Errorf("environ %s must be provided", powerDNSSubdomainAddressEnvKey)
		os.Exit(1)
	}

	usecases, err := newUsecases(conn, fallbackImage, subdomainAddr, e.Logger)
	if err != nil {
		e.Logger.Errorf("failed to initialize handlers: %v", err)
		os.Exit(1)
	}
	// セッション・ルーティング・エラーレスポンスの設定
	handler.Setup(e, secret, usecases, fallbackImage)

	// HTTPサーバ起動
	listenAddr := net.JoinHostPort("", strconv.Itoa(listenPort))
	if err := e.Start(listenAddr); err != nil {
		e.Logger.Errorf("failed to start HTTP server: %v", err)
		os.Exit(1)
	}
}

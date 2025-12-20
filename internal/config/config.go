package config

import (
	"flag"
	_ "github.com/gobugger/gomarket/internal/log"
	"github.com/gobugger/gomarket/internal/util"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	InvoicePaymentWindow    = time.Hour * 6
	OrderProcessingWindow   = time.Hour * 24 * 2
	OrderDispatchWindow     = time.Hour * 24 * 3
	OrderDeliveryWindow     = time.Hour * 24 * 7
	ExtendUnavailableWindow = time.Hour * 24 * 5
)

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
}

var (
	Addr                 string
	DSN                  string
	MoneropayURL         string
	MoneropayCallbackURL string
	RootDir              string
	CssDir               string
	StaticDir            string
	CryptoDir            string
	UploadsBucket        string
	CsrfAuthKey          string
	PgpKey               string
	OnionAddr            string
	SiteName             string
	Socks5Hostname       string
	devMode              bool
	captchaEnabled       bool = true
	entryGuardEnabled    bool
	Minio                MinioConfig
	Cryptocurrency       string
)

// Check it this way so DevMode won't be accidentally set to true
func DevMode() bool {
	return devMode
}

func CaptchaEnabled() bool {
	return captchaEnabled
}

func EntryGuardEnabled() bool {
	return entryGuardEnabled
}

func ParseAndLoad() {
	DefineGlobal()
	Parse()
	Load()
}

func DefineGlobal() {
	flag.StringVar(&Addr, "addr", "localhost:4000", "http address for server to listen")
	flag.StringVar(&DSN, "dsn", os.Getenv("DSN"), "postgres data source name")
	flag.StringVar(&MoneropayURL, "moneropay-url", os.Getenv("MONEROPAY_URL"), "moneropay url")
	flag.StringVar(&MoneropayCallbackURL, "moneropay-cb-url", "localhost:4001", "address for moneropay callbacks")
	flag.StringVar(&RootDir, "root-dir", "./", "root directory for all application data")
	flag.StringVar(&CsrfAuthKey, "csrf-auth-key", os.Getenv("CSRF_AUTH_KEY"), "32 byte csrf auth key")
	flag.StringVar(&OnionAddr, "onion-address", os.Getenv("ONION_ADDRESS"), "onion address")
	flag.StringVar(&SiteName, "name", "GoMarket", "name for this site")
	flag.StringVar(&Cryptocurrency, "cryptocurrency", "XMR", "Cryptocurrency to use [XMR, NANO]")
	flag.BoolVar(&devMode, "dev", false, "set true to enable development mode")
	flag.BoolVar(&captchaEnabled, "captcha", true, "enable captcha")
	flag.BoolVar(&entryGuardEnabled, "entry-guard", true, "enable entry guard")
	flag.StringVar(&Socks5Hostname, "socks5-hostname", "", "set to socks5 proxy hostname (socks5://localhost:9050), defaults to no proxy")
	flag.StringVar(&Minio.Endpoint, "minio-endpoint", os.Getenv("MINIO_ENDPOINT"), "minio endpoint")
	flag.StringVar(&Minio.AccessKeyID, "minio-ak-id", os.Getenv("MINIO_AK_ID"), "minio access key ID")
	flag.StringVar(&Minio.SecretAccessKey, "minio-sa-key", os.Getenv("MINIO_SA_KEY"), "minio secret access key")
}

func Parse() {
	flag.Parse()

	CssDir = "ui/css"
	StaticDir = "static"
	CryptoDir = filepath.Join(StaticDir, "crypto")

	if DevMode() {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))
		slog.SetDefault(logger)
	}
}

func Load() {
	content, err := util.ReadFile(RootDir, filepath.Join(CryptoDir, "pgp.txt"))
	if err != nil {
		slog.Error("failed to read pgp.txt", slog.Any("error", err))
	}
	PgpKey = string(content)
}

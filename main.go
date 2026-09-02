package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
	_ "time/tzdata"

	"github.com/joho/godotenv"

	"github.com/COMTOP1/AFC-GO/infrastructure/telemetry"
	"github.com/COMTOP1/AFC-GO/views"
)

var (
	Version = "unknown"
	Commit  = "unknown"
)

func main() {
	var local, global bool
	var err error
	err = godotenv.Load(".env") // Load .env
	global = err == nil

	err = godotenv.Overload(".env.local") // Load .env.local
	local = err == nil

	fmt.Printf("Version: %s\nCommit: %s\n", Version, Commit)

	otelServiceName := os.Getenv("OTEL_SERVICE_NAME")
	if otelServiceName == "" {
		otelServiceName = "afc-go"
	}

	ctx := context.Background()
	otelShutdown, err := telemetry.Setup(ctx, telemetry.Config{
		ServiceName:    otelServiceName,
		ServiceVersion: Version,
		Endpoint:       os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Headers:        telemetry.ParseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")),
	})
	if err != nil {
		slog.Error(fmt.Sprintf("failed to set up telemetry: %+v", err))
		os.Exit(1)
	}
	defer func() {
		if shutdownErr := otelShutdown(context.Background()); shutdownErr != nil {
			slog.Error(fmt.Sprintf("failed to shut down telemetry: %+v", shutdownErr))
		}
	}()

	// fatal logs msg, flushes telemetry, and exits. Deferred cleanup is
	// skipped by os.Exit, so telemetry must be flushed explicitly here.
	fatal := func(msg string) {
		slog.Error(msg)
		_ = otelShutdown(context.Background())
		os.Exit(1)
	}

	dbHost := os.Getenv("DB_HOSTNAME")
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSL := os.Getenv("DB_SSLMODE")

	if !local && !global && dbHost == "" {
		fatal("unable to find env files and no env variables have been supplied")
	}
	//nolint:gocritic
	if !local && !global {
		slog.Info("using env variables")
	} else if local && global {
		slog.Info("using global and local env files")
	} else if !local {
		slog.Info("using global env file")
	} else {
		slog.Info("using local env file")
	}

	sessionCookieName := os.Getenv("WAUTH_SESSION_COOKIE_NAME")
	if sessionCookieName == "" {
		sessionCookieName = "session"
	}

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		fatal(fmt.Sprintf("invalid option for dbPort: %+v", err))
	}

	dbConnectionString := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
		dbHost,
		dbPort,
		dbUser,
		dbName,
		dbSSL,
		dbPass,
	)

	address := os.Getenv("ADDRESS")

	iter, err := strconv.Atoi(os.Getenv("ITERATIONS"))
	if err != nil {
		fatal(fmt.Sprintf("invalid option for iterations: %+v", err))
	}

	sWorkFactor, err := strconv.Atoi(os.Getenv("SCRYPT_WORK_FACTOR"))
	if err != nil {
		fatal(fmt.Sprintf("invalid option for scrypt work factor: %+v", err))
	}

	sBlockSize, err := strconv.Atoi(os.Getenv("SCRYPT_BLOCK_SIZE"))
	if err != nil {
		fatal(fmt.Sprintf("invalid option for scrypt block size: %+v", err))
	}

	sParallelismFactor, err := strconv.Atoi(os.Getenv("SCRYPT_PARALLELISM_FACTOR"))
	if err != nil {
		fatal(fmt.Sprintf("invalid option for scrypt parallelism factor: %+v", err))
	}

	keyLen, err := strconv.Atoi(os.Getenv("KEY_LENGTH_BYTES"))
	if err != nil {
		fatal(fmt.Sprintf("invalid option for key length: %+v", err))
	}

	var fileDir string

	stat, err := os.Stat("/FileStore")
	if err == nil && stat.IsDir() {
		slog.Info("using root /FileStore")
		fileDir = "/FileStore"
	} else {
		stat, err = os.Stat("./FileStore")
		if err == nil && stat.IsDir() {
			slog.Info("using local ./FileStore")
			fileDir = "./FileStore"
		} else {
			fatal(fmt.Sprintf("failed to get fileStore - stat: %+v, error: %+v", stat, err))
		}
	}

	mailPort, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))

	domainName := os.Getenv("DOMAIN_NAME")

	// Generate config
	conf := &views.Config{
		Address:           address,
		DatabaseURL:       dbConnectionString,
		DomainName:        domainName,
		SessionCookieName: sessionCookieName,
		FileDir:           fileDir,
		Mail: views.SMTPConfig{
			Host:     os.Getenv("MAIL_HOST"),
			Username: os.Getenv("MAIL_USER"),
			Password: os.Getenv("MAIL_PASS"),
			Port:     mailPort,
		},
		Security: views.SecurityConfig{
			EncryptionKey:           os.Getenv("ENCRYPTION_KEY"),
			AuthenticationKey:       os.Getenv("AUTHENTICATION_KEY"),
			Iterations:              iter,
			ScryptWorkFactor:        sWorkFactor,
			ScryptBlockSize:         sBlockSize,
			ScryptParallelismFactor: sParallelismFactor,
			KeyLength:               keyLen,
		},
	}

	v := views.New(conf, dbHost, 30*time.Second)

	router := NewRouter(&RouterConf{
		Config: conf,
		Views:  v,
	})

	err = router.Start()
	v.Stop()
	fatal(fmt.Sprintf("The web server couldn't be started!\n\n%s\n\nExiting!", err))
}

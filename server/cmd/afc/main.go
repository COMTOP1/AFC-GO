// Command afc serves the AFC Aldermaston website and its JSON API.
//
//	@title			AFC Aldermaston API
//	@version		1
//	@description	JSON API behind the AFC Aldermaston website.
//	@BasePath		/api/v1
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	_ "time/tzdata"

	"github.com/joho/godotenv"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/telemetry"
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

	mailPort, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))

	domainName := os.Getenv("DOMAIN_NAME")
	if domainName == "" {
		slog.Warn("DOMAIN_NAME is empty: the session cookie will be marked Secure, so logging in over plain HTTP will not work; set DOMAIN_NAME (or have it start with \"localhost\" for local development over HTTP)")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}

	// Redis/Valkey is optional - when REDIS_ADDRESSES isn't set, an
	// in-process cache is used instead (fine for a single instance, but
	// state such as password reset tokens won't be shared if you run more
	// than one instance).
	var redisAddresses []string
	if raw := os.Getenv("REDIS_ADDRESSES"); raw != "" {
		for _, address := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(address); trimmed != "" {
				redisAddresses = append(redisAddresses, trimmed)
			}
		}
	}
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisTLS, _ := strconv.ParseBool(os.Getenv("REDIS_TLS"))

	a := app.New(app.Config{
		Address:      address,
		DomainName:   domainName,
		DatabaseURL:  dbConnectionString,
		DatabaseHost: dbHost,
		Version:      Version,
		Session: auth.Config{
			CookieName:        sessionCookieName,
			AuthenticationKey: os.Getenv("AUTHENTICATION_KEY"),
			EncryptionKey:     os.Getenv("ENCRYPTION_KEY"),
			Secure:            !strings.HasPrefix(domainName, "localhost"),
		},
		S3: storage.Config{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Region:    s3Region,
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
		},
		Mail: mail.Config{
			Host:     os.Getenv("MAIL_HOST"),
			Username: os.Getenv("MAIL_USER"),
			Password: os.Getenv("MAIL_PASS"),
			Port:     mailPort,
		},
		Passwords: auth.PasswordConfig{
			Iterations:              iter,
			ScryptWorkFactor:        sWorkFactor,
			ScryptBlockSize:         sBlockSize,
			ScryptParallelismFactor: sParallelismFactor,
			KeyLength:               keyLen,
		},
		Redis: auth.RedisConfig{
			Addresses:  redisAddresses,
			MasterName: os.Getenv("REDIS_MASTER_NAME"),
			Username:   os.Getenv("REDIS_USERNAME"),
			Password:   os.Getenv("REDIS_PASSWORD"),
			DB:         redisDB,
			TLS:        redisTLS,
			KeyPrefix:  os.Getenv("REDIS_KEY_PREFIX"),
		},
	})

	err = a.Start()
	a.Stop()
	fatal(fmt.Sprintf("The web server couldn't be started!\n\n%s\n\nExiting!", err))
}

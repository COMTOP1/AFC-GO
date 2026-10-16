// Command resetlegacypasswords removes the old site's unsalted SHA-512
// password digests and emails each affected user a link to set a new
// password. Run it once against production; running it again does nothing.
//
// It needs the same database, Redis, mail and DOMAIN_NAME settings as the
// server. Redis is required: reset links must be stored where the server
// can find them, and without Redis each process keeps its own.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // postgres driver

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/emails"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Overload(".env.local")
	os.Exit(run(context.Background()))
}

func run(ctx context.Context) int {
	var redisAddresses []string
	for _, address := range strings.Split(os.Getenv("REDIS_ADDRESSES"), ",") {
		if trimmed := strings.TrimSpace(address); trimmed != "" {
			redisAddresses = append(redisAddresses, trimmed)
		}
	}
	if len(redisAddresses) == 0 {
		slog.Error("REDIS_ADDRESSES is required: without Redis the server can't see the reset links this command issues")
		return 1
	}
	domain := os.Getenv("DOMAIN_NAME")
	if domain == "" {
		slog.Error("DOMAIN_NAME is required to build reset links")
		return 1
	}
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		slog.Error(fmt.Sprintf("invalid option for DB_PORT: %+v", err))
		return 1
	}
	mailPort, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisTLS, _ := strconv.ParseBool(os.Getenv("REDIS_TLS"))

	conn, err := sqlx.ConnectContext(ctx, "postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
		os.Getenv("DB_HOSTNAME"), dbPort, os.Getenv("DB_USERNAME"), os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"), os.Getenv("DB_PASSWORD")))
	if err != nil {
		slog.Error(fmt.Sprintf("failed to connect to the database: %+v", err))
		return 1
	}
	defer conn.Close()

	tokens := auth.NewTokens(auth.RedisConfig{
		Addresses:  redisAddresses,
		MasterName: os.Getenv("REDIS_MASTER_NAME"),
		Username:   os.Getenv("REDIS_USERNAME"),
		Password:   os.Getenv("REDIS_PASSWORD"),
		DB:         redisDB,
		TLS:        redisTLS,
		KeyPrefix:  os.Getenv("REDIS_KEY_PREFIX"),
	})
	defer tokens.Close()
	sender := emails.NewSMTP(mail.NewMailer(mail.Config{
		Host:     os.Getenv("MAIL_HOST"),
		Username: os.Getenv("MAIL_USER"),
		Password: os.Getenv("MAIL_PASS"),
		Port:     mailPort,
	}))

	store := user.NewUserRepo(conn)
	// Only ResetPassword is used, which needs neither teams nor file storage.
	svc := user.NewService(store, nil, nil, sender, tokens, user.HashParams{}, domain)
	return resetAll(ctx, store, svc)
}

type resetter interface {
	ResetPassword(ctx context.Context, id int) (user.ResetResult, error)
}

func resetAll(ctx context.Context, store *user.Store, svc resetter) int {
	ids, err := store.ClearLegacyPasswords(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to clear legacy passwords: %+v", err))
		return 1
	}
	var unsent []int
	for _, id := range ids {
		res, resetErr := svc.ResetPassword(ctx, id)
		if resetErr != nil || !res.EmailSent {
			unsent = append(unsent, id)
		}
	}
	slog.Info("cleared legacy passwords", "count", len(ids), "emailed", len(ids)-len(unsent))
	if len(unsent) > 0 {
		// The accounts are already cleared and flagged; an admin can send
		// each of these a new link with "Reset password" on /users.
		slog.Error("reset links not emailed; resend them from /users", "userIDs", unsent)
		return 1
	}
	return 0
}

// Command wraplegacypasswords wraps the old site's unsalted SHA-512 password
// digests in scrypt, for accounts that haven't logged in since the move to
// this server. Users keep their passwords. Run it once against production;
// running it again does nothing.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // postgres driver

	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Overload(".env.local")
	os.Exit(run(context.Background()))
}

func run(ctx context.Context) int {
	params := map[string]int{}
	for _, name := range []string{"SCRYPT_WORK_FACTOR", "SCRYPT_BLOCK_SIZE", "SCRYPT_PARALLELISM_FACTOR", "KEY_LENGTH_BYTES"} {
		v, err := strconv.Atoi(os.Getenv(name))
		if err != nil {
			slog.Error(fmt.Sprintf("invalid option for %s: %+v", name, err))
			return 1
		}
		params[name] = v
	}
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		slog.Error(fmt.Sprintf("invalid option for DB_PORT: %+v", err))
		return 1
	}
	conn, err := sqlx.ConnectContext(ctx, "postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
		os.Getenv("DB_HOSTNAME"), dbPort, os.Getenv("DB_USERNAME"), os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"), os.Getenv("DB_PASSWORD")))
	if err != nil {
		slog.Error(fmt.Sprintf("failed to connect to the database: %+v", err))
		return 1
	}
	defer conn.Close()

	n, err := user.NewUserRepo(conn).WrapLegacyPasswords(ctx, params["SCRYPT_WORK_FACTOR"], params["SCRYPT_BLOCK_SIZE"],
		params["SCRYPT_PARALLELISM_FACTOR"], params["KEY_LENGTH_BYTES"])
	if err != nil {
		slog.Error("failed to wrap legacy passwords", "wrapped", n, "error", err)
		return 1
	}
	slog.Info("wrapped legacy passwords", "count", n)
	return 0
}

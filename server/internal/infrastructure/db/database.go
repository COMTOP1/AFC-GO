package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

// NewStore initialises the store
func NewStore(dataSourceName string, host string) *sqlx.DB {
	db, err := sqlx.ConnectContext(context.Background(), "postgres", dataSourceName)
	if err != nil {
		slog.Error(fmt.Sprintf("db failed: %+v", err))
		os.Exit(1)
	}
	slog.Info("connected to db: " + host)
	return db
}

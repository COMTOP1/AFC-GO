package db

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

// NewStore initialises the store
func NewStore(dataSourceName string, host string) *sqlx.DB {
	db, err := sqlx.ConnectContext(context.Background(), "postgres", dataSourceName)
	if err != nil {
		log.Fatalf("db failed: %+v", err)
	}
	log.Printf("connected to db: %s", host)
	return db
}

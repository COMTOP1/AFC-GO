package db

import (
	"context"
	"log"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/jmoiron/sqlx"
)

// NewStore initialises the store
func NewStore(dataSourceName string, host string) *sqlx.DB {
	db, err := sqlx.ConnectContext(context.Background(), "mysql", dataSourceName)
	if err != nil {
		log.Fatalf("db failed: %+v", err)
	}
	log.Printf("connected to db: %s", host)
	return db
}

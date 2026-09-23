package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	var local, global bool
	var err error
	err = godotenv.Load(".env") // Load .env
	global = err == nil

	err = godotenv.Overload(".env.local") // Load .env.local
	local = err == nil

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

	address := os.Getenv("ADDRESS")

	domainName := os.Getenv("DOMAIN_NAME")

	// Generate config
	conf := &Config{
		Address:    address,
		DomainName: domainName,
	}

	v := New(conf)

	router := NewRouter(&RouterConf{
		Config: conf,
		Views:  v,
	})

	err = router.Start()
	slog.Error(fmt.Sprintf("The web server couldn't be started!\n\n%s\n\nExiting!", err))
	os.Exit(1)
}

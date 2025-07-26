package main

import (
	"log"
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
		log.Println("using env variables")
	} else if local && global {
		log.Println("using global and local env files")
	} else if !local {
		log.Println("using global env file")
	} else {
		log.Println("using local env file")
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
	log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
}

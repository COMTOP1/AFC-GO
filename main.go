package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/COMTOP1/AFC-GO/views"
)

func main() {
	var local, global bool
	var err error
	err = godotenv.Load(".env") // Load .env
	global = err == nil

	err = godotenv.Overload(".env.local") // Load .env.local
	local = err == nil

	dbHost := os.Getenv("DB_HOSTNAME")
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if !local && !global && dbHost == "" {
		log.Fatal("unable to find env files and no env variables have been supplied")
	}
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

	sessionCookieName := os.Getenv("WAUTH_SESSION_COOKIE_NAME")
	if sessionCookieName == "" {
		sessionCookieName = "session"
	}

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("invalid option for dbPort: %+v", err)
	}

	dbConnectionString := fmt.Sprintf(
		"%s:%s@(%s:%d)/%s?parseTime=true",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)

	address := os.Getenv("ADDRESS")

	iter, err := strconv.Atoi(os.Getenv("ITERATIONS"))
	if err != nil {
		log.Fatalf("invalid option for iterations: %+v", err)
	}

	keyLen, err := strconv.Atoi(os.Getenv("KEY_LENGTH_BYTES"))
	if err != nil {
		log.Fatalf("invalid option for key length: %+v", err)
	}

	var fileDir string

	stat, err := os.Stat("/FileStore")
	if err == nil && stat.IsDir() {
		log.Println("using root /FileStore")
		fileDir = "/FileStore"
	} else {
		stat, err = os.Stat("./FileStore")
		if err == nil && stat.IsDir() {
			log.Println("using local ./FileStore")
			fileDir = "./FileStore"
		} else {
			log.Fatalf("failed to get fileStore - stat: %+v, error: %+v", stat, err)
		}
	}

	// Generate config
	conf := &views.Config{
		Address:     address,
		DatabaseURL: dbConnectionString,
		// BaseDomainName:    os.Getenv("WAUTH_BASE_DOMAIN_NAME"),
		// DomainName:        domainName,
		// LogoutEndpoint:    os.Getenv("WAUTH_LOGOUT_ENDPOINT"),
		SessionCookieName: sessionCookieName,
		FileDir:           fileDir,
		// Mail: views.SMTPConfig{
		//	Host:       os.Getenv("WAUTH_MAIL_HOST"),
		//	Username:   os.Getenv("WAUTH_MAIL_USER"),
		//	Password:   os.Getenv("WAUTH_MAIL_PASS"),
		//	Port:       mailPort,
		//	DomainName: domainName,
		// },
		Security: views.SecurityConfig{
			EncryptionKey:     os.Getenv("ENCRYPTION_KEY"),
			AuthenticationKey: os.Getenv("AUTHENTICATION_KEY"),
			//SigningKey:        signingKey,
			Iterations: iter,
			KeyLength:  keyLen,
		},
	}

	v := views.New(conf, dbHost)

	router := NewRouter(&RouterConf{
		Config: conf,
		Views:  v,
	})

	err = router.Start()
	if err != nil {
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
}

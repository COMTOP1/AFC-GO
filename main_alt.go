package main

import (
	"fmt"
	"log"
	"os"

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

	signingKey := os.Getenv("WAUTH_SIGNING_KEY")
	dbHost := os.Getenv("WAUTH_DB_HOST")

	if !local && !global && signingKey == "" && dbHost == "" {
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

	dbPort := os.Getenv("WAUTH_DB_PORT")

	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		dbHost,
		dbPort,
		os.Getenv("WAUTH_DB_USER"),
		os.Getenv("WAUTH_DB_NAME"),
		os.Getenv("WAUTH_DB_SSLMODE"),
		os.Getenv("WAUTH_DB_PASS"),
	)

	//mailPort, _ := strconv.Atoi(os.Getenv("WAUTH_MAIL_PORT"))

	address := os.Getenv("WAUTH_ADDRESS")

	domainName := os.Getenv("WAUTH_DOMAIN_NAME")

	//adPort, err := strconv.Atoi(os.Getenv("WAUTH_AD_PORT"))
	//if err != nil {
	//	log.Fatalf("failed to get ad port env: %+v", err)
	//}
	//
	//adSecurity, err := strconv.Atoi(os.Getenv("WAUTH_AD_SECURITY"))
	//if err != nil {
	//	log.Fatalf("failed to get ad security env: %+v", err)
	//}

	// Generate config
	conf := &views.Config{
		Address:           address,
		DatabaseURL:       dbConnectionString,
		BaseDomainName:    os.Getenv("WAUTH_BASE_DOMAIN_NAME"),
		DomainName:        domainName,
		LogoutEndpoint:    os.Getenv("WAUTH_LOGOUT_ENDPOINT"),
		SessionCookieName: sessionCookieName,
		//Mail: views.SMTPConfig{
		//	Host:       os.Getenv("WAUTH_MAIL_HOST"),
		//	Username:   os.Getenv("WAUTH_MAIL_USER"),
		//	Password:   os.Getenv("WAUTH_MAIL_PASS"),
		//	Port:       mailPort,
		//	DomainName: domainName,
		//},
		Security: views.SecurityConfig{
			EncryptionKey:     os.Getenv("WAUTH_ENCRYPTION_KEY"),
			AuthenticationKey: os.Getenv("WAUTH_AUTHENTICATION_KEY"),
			SigningKey:        signingKey,
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

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/polarpayne/pp"
)

func EnvDef(key, def string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return value
}

func main() {
	var (
		flagDBNoInit          = flag.Bool("db-no-init", false, "do not run create table and migrations on the database before starting")
		flagDBConn            = flag.String("db-conn", EnvDef("DB_CONN", "postgres://pp:secret@localhost/pp?sslmode=disable"), "connection string to connect to db")
		flagOAuthClientID     = flag.String("oauth-client-id", os.Getenv("OAUTH_CLIENT_ID"), "OAuth2 Client ID that is used for Google SSO")
		flagOAuthClientSecret = flag.String("oauth-client-secret", os.Getenv("OAUTH_CLIENT_SECRET"), "OAuth2 Client Secret that is used for Google SSO")
		flagBackendBucket     = flag.String("backend-bucket", os.Getenv("BACKEND_BUCKET"), "name of the bucket that stores the podcasts")
		flagBackendLogo       = flag.String("backend-logo", EnvDef("BACKEND_LOGO", "logo.png"), "key of the logo within the backend bucket")
		flagBaseURL           = flag.String("base-url", EnvDef("BASE_URL", "http://localhost:8080"), "base URL of the application, used to generate correct URLs")
		flagHost              = flag.String("host", EnvDef("HOST", "localhost"), "address the application should bind to")
		flagPort              = flag.String("port", EnvDef("PORT", "8080"), "port that the application will listen to")
		flagName              = flag.String("name", EnvDef("PODCAST_NAME", "Unnamed Podcast"), "name of the podcast")
		flagDescription       = flag.String("description", EnvDef("PODCAST_DESCRIPTION", "No Description"), "description of the podcast")
		flagHelpText          = flag.String("help-text", os.Getenv("HELP_TEXT"), "help text that is shown at the bottom of the homepage")
	)

	flag.Parse()

	if strings.HasSuffix(*flagBaseURL, "/") {
		log.Fatalf("base-url must not end with a '/' (it was %q)", *flagBaseURL)
	}

	herokuDatabaseURL := os.Getenv("DATABASE_URL")
	if herokuDatabaseURL != "" {
		log.Print("DATABASE_URL was set (we are likely running in Heroku), overriding db-conn flag with it")
		*flagDBConn = herokuDatabaseURL
	}

	backend := pp.NewBackendS3(*flagBackendBucket, *flagBackendLogo)
	auth := pp.NewAuthGoogle(*flagOAuthClientID, *flagOAuthClientSecret, *flagBaseURL+"/auth")
	storage, err := pp.NewStoragePostgres(*flagDBConn)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	if !*flagDBNoInit {
		err := storage.Init()
		if err != nil {
			log.Fatalf("failed to init storage: %v", err)
		}
	}

	addr := net.JoinHostPort(*flagHost, *flagPort)

	s := newServer(
		*flagBaseURL, *flagName, *flagDescription, *flagHelpText,
		backend, auth, storage,
	)
	log.Fatal(s.start(addr, 5*time.Minute))
}

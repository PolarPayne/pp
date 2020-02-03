package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/lib/pq"
	"github.com/polarpayne/pp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func EnvDef(key, def string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return value
}

var (
	flagDBNoInit          = flag.Bool("db-no-init", false, "do not run create table and migrations on the database before starting")
	flagDB                = flag.String("db-engine", EnvDef("DB_ENGINE", "postgres"), "db engine that will be used (currently only postgres is supported)")
	flagDBConn            = flag.String("db-conn", EnvDef("DB_CONN", "postgres://pp:secret@localhost/pp?sslmode=disable"), "connection string to connect to db")
	flagOAuthClientID     = flag.String("oauth-client-id", os.Getenv("OAUTH_CLIENT_ID"), "OAuth2 Client ID that is used for Google SSO")
	flagOAuthClientSecret = flag.String("oauth-client-secret", os.Getenv("OAUTH_CLIENT_SECRET"), "OAuth2 Client Secret that is used for Google SSO")
	flagBackendBucket     = flag.String("backend-bucket", os.Getenv("BACKEND_BUCKET"), "name of the bucket that stores the podcasts")
	flagBaseURL           = flag.String("base-url", EnvDef("BASE_URL", "http://localhost:8080"), "base URL of the application, used to generate correct URLs")
	flagAddr              = flag.String("addr", EnvDef("ADDR", ":8080"), "address that the application will bind to")
	flagName              = flag.String("name", EnvDef("PODCAST_NAME", "Unnamed Podcast"), "name of the podcast")
	flagDescription       = flag.String("description", EnvDef("PODCAST_DESCRIPTION", "No Description"), "description of the podcast")
	flagHelpText          = flag.String("help-text", os.Getenv("HELP_TEXT"), "help text that is shown at the bottom of the homepage")
)

func main() {
	flag.Parse()

	if strings.HasSuffix(*flagBaseURL, "/") {
		log.Fatalf("base-url must not end with a '/' (it was %q)", *flagBaseURL)
	}

	herokuDatabaseURL := os.Getenv("DATABASE_URL")
	if herokuDatabaseURL != "" {
		log.Print("DATABASE_URL was set (we are likely running in Heroku), overriding db-conn flag with it")
		*flagDBConn = herokuDatabaseURL
	}

	db, err := sql.Open(*flagDB, *flagDBConn)
	if err != nil {
		log.Fatalf("failed to open DB connection: %v", err)
	}

	if !*flagDBNoInit {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
			user_id    TEXT UNIQUE NOT NULL,
			secret     TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`)
		if err != nil {
			log.Fatalf("failed to create table(s) in db: %v", err)
		}

		log.Print("created tables in the database")
	}

	session := session.Must(session.NewSession())
	oauth := &oauth2.Config{
		ClientID:     *flagOAuthClientID,
		ClientSecret: *flagOAuthClientSecret,
		RedirectURL:  *flagBaseURL + "/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	s := newServer(*flagBaseURL, *flagName, *flagDescription, *flagHelpText, pp.NewBackend(session, *flagBackendBucket), oauth, db)
	log.Fatal(s.start(":8080", 15*time.Second))
}

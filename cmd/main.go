package main

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/polarpayne/pp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	session := session.Must(session.NewSession())
	oauth := &oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	s := newServer("http://localhost:8080", "Reaktor", "Reaktor Podcast", pp.NewBackendS3(session, "reaktor-podcasts"), oauth)
	log.Fatal(s.start(":8080", 15*time.Second))
}

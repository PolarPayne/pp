package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

func (s *server) handleAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		code := q.Get("code")

		if code == "" {
			log.Printf("auth code is not set, redirecting user to authentication")
			url := s.oauth.AuthCodeURL("state")
			http.Redirect(w, r, url, 302)
			return
		}

		tok, err := s.oauth.Exchange(r.Context(), code)
		if err != nil {
			log.Printf("failed to exchange token: %v", err)
			s.handleError(w, r, err)
			return
		}
		spew.Dump(tok)

		client := s.oauth.Client(r.Context(), tok)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			log.Printf("failed to get email: %v", err)
			s.handleError(w, r, err)
			return
		}
		defer resp.Body.Close()

		res := new(bytes.Buffer)
		_, err = io.Copy(res, resp.Body)
		if err != nil {
			s.handleError(w, r, err)
			return
		}
		spew.Dump(res)

		http.SetCookie(w, &http.Cookie{Name: "podcast_session", Value: "test"})
		http.Redirect(w, r, "/", 302)
	}
}

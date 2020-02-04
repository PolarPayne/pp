package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type googleUserinfo struct {
	ID    string
	Email string
}

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

		client := s.oauth.Client(r.Context(), tok)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			log.Printf("failed to get email: %v", err)
			s.handleError(w, r, err)
			return
		}
		defer resp.Body.Close()

		jDecoder := json.NewDecoder(resp.Body)
		userinfo := googleUserinfo{}
		err = jDecoder.Decode(&userinfo)
		if err != nil {
			s.handleError(w, r, err)
			log.Panicf("failed to parse JSON returned by Google: %v", err)
			return
		}

		email := userinfo.Email
		secret, err := s.createUser(email)
		if err != nil {
			s.handleError(w, r, err)
			log.Panicf("failed to create a user: %v", err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "podcast_session",
			Value:    secret,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, r, "/", 302)
	}
}

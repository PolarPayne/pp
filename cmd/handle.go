package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/eduncan911/podcast"
)

// handleError writes the error message to the log and then sends a 500 status code
// this can (and does) generate additional log messages because there is no way
// to check if WriteHeader has already been called.
// These log messages look like `http: superfluous response.WriteHeader call from`,
// and are completely harmless.
func (s *server) handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error when handling a request to %q: %v", r.URL.EscapedPath(), err)
	w.WriteHeader(http.StatusInternalServerError)
}

func (s *server) handleHTTPToHTTPS(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		forwardedProto := r.Header.Get("X-Forwarded-Proto")
		if forwardedProto == "http" && strings.HasPrefix(s.baseURL, "https") {
			log.Printf("request with X-Forwarded-Proto equal to HTTP and base URL is a HTTPS URL, redirecting user to HTTPS")
			url := s.baseURL + r.URL.String()
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		f(w, r)
	}
}

// handleSecret returns true if and only if the secret is valid, if handleSecret returns false
// it will also write the correct status code and message to the response.
func (s *server) handleSecret(w http.ResponseWriter, r *http.Request) (string, bool) {
	q := r.URL.Query()
	secret := q.Get("s")

	// TODO: logo access to database

	ok, err := s.storage.ValidSecret(secret)
	if err != nil {
		s.handleError(w, r, err)
		return "", false
	}
	if !ok {
		log.Printf("invalid secret %q when trying to access feed", secret)
		w.WriteHeader(http.StatusForbidden)
		return "", false
	}

	return secret, true
}

func (s *server) handleLogo(w http.ResponseWriter, r *http.Request) {
	logo, err := s.backend.GetLogo()
	if err != nil {
		s.handleError(w, r, err)
		return
	}
	defer logo.Close()

	_, err = io.Copy(w, logo)
	if err != nil {
		log.Printf("failed to write logo to response: %v", err)
	}
}

func (s *server) handleAuth(w http.ResponseWriter, r *http.Request) {
	err := s.auth.HandleAuth(w, r, s.storage, *flagNoSecureCookie)
	if err != nil {
		s.handleError(w, r, err)
	}
}

func (s *server) handleFeed(w http.ResponseWriter, r *http.Request) {
	secret, ok := s.handleSecret(w, r)
	if !ok {
		return
	}

	err := s.storage.LogFeed(secret, r.Referer(), r.UserAgent())
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	now := time.Now()

	feed := podcast.New(s.name, s.baseURL, s.description, nil, &now)
	feed.IBlock = "yes"
	feed.Image = &podcast.Image{URL: s.baseURL + "/logo"}
	for _, p := range s.getPodcasts() {
		pd := p.Details()

		q := url.Values{}
		q.Set("s", secret)
		q.Set("n", pd.Key)

		description := pd.Description
		if description == "" {
			description = pd.Title
		}

		feed.AddItem(podcast.Item{
			Title:       pd.Title,
			Description: description,
			PubDate:     &pd.Published,
			Enclosure: &podcast.Enclosure{
				Length: pd.Size,
				Type:   podcast.MP3,
				URL:    s.baseURL + "/podcast?" + q.Encode(),
			},
		})
	}

	err = feed.Encode(w)
	if err != nil {
		log.Printf("failed to write feed to response: %v", err)
	}
}

func (s *server) handlePodcast(w http.ResponseWriter, r *http.Request) {
	secret, ok := s.handleSecret(w, r)
	if !ok {
		return
	}

	name := r.URL.Query().Get("n")

	err := s.storage.LogPodcast(secret, name, r.Referer(), r.UserAgent())
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	for _, podcast := range s.getPodcasts() {
		if podcast.Details().Key == name {
			err := podcast.HandlePodcast(w, r)
			if err != nil {
				s.handleError(w, r, err)
			}
			return
		}
	}

	// 403 would already be written if the user doesn't have access
	// therefore we aren't leaking any information
	w.WriteHeader(http.StatusNotFound)
}

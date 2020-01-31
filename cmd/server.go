package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/polarpayne/pp"
	"golang.org/x/oauth2"
)

func logHandle(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request: url=%q", r.URL)
		h(w, r)
	}
}

type server struct {
	mux *http.ServeMux

	baseURL     string
	name        string
	description string
	backend     pp.Backend
	oauth       *oauth2.Config

	podcasts      []pp.Podcast
	podcastsMutex sync.RWMutex
}

func newServer(baseURL, name, description string, backend pp.Backend, oauth *oauth2.Config) *server {
	out := new(server)
	out.baseURL = baseURL
	out.name = name
	out.description = description
	out.backend = backend
	out.oauth = oauth

	out.mux = http.NewServeMux()
	out.mux.HandleFunc("/", logHandle(out.handleHome()))
	out.mux.HandleFunc("/auth", logHandle(out.handleAuth()))
	out.mux.HandleFunc("/feed", logHandle(out.handleFeed()))
	out.mux.HandleFunc("/podcast", logHandle(out.handlePodcast()))

	return out
}

func (s *server) updatePodcasts() error {
	log.Printf("updating podcasts")

	ps, err := s.backend.ListPodcasts()
	if err != nil {
		return err
	}

	s.podcastsMutex.Lock()
	defer s.podcastsMutex.Unlock()

	log.Printf("updating podcasts: found %v podcasts", len(ps))
	s.podcasts = ps

	return nil
}

func (s *server) getPodcasts() []pp.Podcast {
	s.podcastsMutex.RLock()
	defer s.podcastsMutex.RUnlock()
	return s.podcasts
}

func (s *server) start(addr string, updateInterval time.Duration) error {
	err := s.updatePodcasts()
	if err != nil {
		return err
	}

	go func() {
		var (
			err      error
			errCount int
		)

		ticker := time.NewTicker(updateInterval)
		for range ticker.C {
			err = s.updatePodcasts()
			if err != nil {
				if errCount >= 3 {
					log.Panicf("failed to updated podcasts >3 times in a row: %v", err)
				}
				log.Printf("failed to update podcasts, error count = %v", errCount)
				errCount++
			} else {
				errCount = 0
			}
		}
	}()

	log.Printf("starting server... listening on %v", addr)
	return http.ListenAndServe(addr, s.mux)
}

func (s *server) handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error when handling a request: %v", r.URL.EscapedPath())
	w.WriteHeader(500)
}

package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/polarpayne/pp"
	"golang.org/x/oauth2"
)

type podcastList []pp.Podcast

func (a podcastList) Len() int      { return len(a) }
func (a podcastList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a podcastList) Less(i, j int) bool {
	if a[i].Published.Equal(a[j].Published) {
		return a[i].Title < a[i].Title
	}
	return a[i].Published.Before(a[j].Published)
}

type server struct {
	mux *http.ServeMux

	baseURL     string
	name        string
	description string
	helpText    string
	backend     *pp.Backend
	oauth       *oauth2.Config
	db          *sql.DB

	podcasts      podcastList
	podcastsMutex sync.RWMutex
}

func newServer(baseURL, name, description, helpText string, backend *pp.Backend, oauth *oauth2.Config, db *sql.DB) *server {
	out := new(server)
	out.baseURL = baseURL
	out.name = name
	out.description = description
	out.helpText = helpText
	out.backend = backend
	out.oauth = oauth
	out.db = db

	out.mux = http.NewServeMux()
	out.mux.HandleFunc("/", out.handleHome())
	out.mux.HandleFunc("/auth", out.handleAuth())
	out.mux.HandleFunc("/feed", out.handleFeed())
	out.mux.HandleFunc("/podcast", out.handlePodcast())

	out.mux.HandleFunc("/logo", out.handleLogo())
	out.mux.HandleFunc("/favicon.ico", out.handleLogo())

	return out
}

func (s *server) handleLogo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	s.podcasts = make([]pp.Podcast, 0, len(ps))

	now := time.Now()
	for _, p := range ps {
		if now.Before(p.Published) {
			log.Printf("updating podcasts: skipping podcast with published date in the future title=%q published=%v", p.Title, p.Published)
			continue
		}
		s.podcasts = append(s.podcasts, p)
	}

	sort.Sort(s.podcasts)

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
	log.Printf("internal server error when handling a request to %q: %v", r.URL.EscapedPath(), err)
	w.WriteHeader(500)
}

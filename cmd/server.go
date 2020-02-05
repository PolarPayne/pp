package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/polarpayne/pp"
)

type server struct {
	mux *http.ServeMux

	baseURL     string
	name        string
	description string
	helpText    string
	backend     pp.Backend
	auth        pp.Auth
	storage     pp.Storage

	podcasts      podcastList
	podcastsMutex sync.RWMutex
}

func newServer(baseURL, name, description, helpText string, backend pp.Backend, auth pp.Auth, storage pp.Storage) *server {
	out := new(server)

	out.baseURL = baseURL
	out.name = name
	out.description = description
	out.helpText = helpText

	out.backend = backend
	out.auth = auth
	out.storage = storage

	out.mux = http.NewServeMux()

	out.mux.HandleFunc("/", out.handleHTTPToHTTPS(out.handleHome()))

	out.mux.HandleFunc("/logo", out.handleHTTPToHTTPS(out.handleLogo))
	out.mux.HandleFunc("/favicon.ico", out.handleHTTPToHTTPS(out.handleLogo))

	out.mux.HandleFunc("/auth", out.handleHTTPToHTTPS(out.handleAuth))

	out.mux.HandleFunc("/feed", out.handleHTTPToHTTPS(out.handleFeed))
	out.mux.HandleFunc("/podcast", out.handleHTTPToHTTPS(out.handlePodcast))

	return out
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

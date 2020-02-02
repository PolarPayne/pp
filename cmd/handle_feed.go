package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/eduncan911/podcast"
)

func (s *server) handleFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		secret := q.Get("s")

		ok, err := s.validSecret(secret)
		if err != nil {
			s.handleError(w, r, err)
			return
		}
		if !ok {
			log.Printf("invalid secret %q when trying to access feed", secret)
			w.WriteHeader(403)
			return
		}

		// TODO: log access to db

		now := time.Now()

		feed := podcast.New(s.name, s.baseURL, s.description, nil, &now)
		for _, p := range s.getPodcasts() {
			q := url.Values{}
			q.Set("s", secret)
			q.Set("n", p.Key)

			feed.AddItem(podcast.Item{
				Title:       p.Title,
				Description: p.Title,
				PubDate:     &p.Published,
				Enclosure: &podcast.Enclosure{
					Length: p.Size,
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
}

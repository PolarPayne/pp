package main

import (
	"log"
	"net/http"
	"time"

	"github.com/eduncan911/podcast"
)

func (s *server) handleFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		secret := q.Get("s")
		_ = secret

		now := time.Now()

		feed := podcast.New(s.name, s.baseURL, s.description, nil, &now)
		for _, p := range s.getPodcasts() {
			title := p.Title()
			published := p.Published()
			size := p.Size()

			feed.AddItem(podcast.Item{
				Title:       title,
				Description: title,
				PubDate:     &published,
				Enclosure: &podcast.Enclosure{
					Length: size,
					Type:   podcast.MP3,
					URL:    s.baseURL + "/podcast?n=2020-01-19+Hello+from+Reaktor.mp3",
				},
			})
		}

		err := feed.Encode(w)
		if err != nil {
			log.Printf("failed to write feed to response: %v", err)
		}
	}
}

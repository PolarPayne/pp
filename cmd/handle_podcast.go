package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
)

func (s *server) handlePodcast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		secret, name := q.Get("s"), q.Get("n")
		log.Printf("podcast secret=%q name=%q", secret, name)

		p, err := s.backend.GetPodcast(name)
		if err != nil {
			s.handleError(w, r, err)
			return
		}

		pc, err := p.Open()
		if err != nil {
			s.handleError(w, r, err)
			return
		}
		defer pc.Close()

		size := p.Size()

		w.Header().Add("Content-Type", "audio/mpeg")
		w.Header().Add("Content-Length", strconv.FormatInt(size, 10))
		n, err := io.Copy(w, pc)
		if err != nil {
			log.Printf("failed to copy content to response (%v of %v bytes copied): %v", n, size, err)
		}
	}
}

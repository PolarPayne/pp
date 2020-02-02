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

		// check the secret is valid
		ok, err := s.validSecret(secret)
		if err != nil {
			s.handleError(w, r, err)
			return
		}
		if !ok {
			w.WriteHeader(403)
			return
		}

		// get podcast and stream it to the client
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

		w.Header().Add("Content-Type", "audio/mpeg")
		w.Header().Add("Content-Length", strconv.FormatInt(p.Size, 10))
		n, err := io.Copy(w, pc)
		if err != nil {
			log.Printf("failed to copy content to response (%v of %v bytes copied): %v", n, p.Size, err)
		}
	}
}
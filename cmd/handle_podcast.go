package main

import (
	"net/http"
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

		err = p.HandleHTTP(w, r)
	}
}

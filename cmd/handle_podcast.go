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

		for _, podcast := range s.getPodcasts() {
			if podcast.Key == name {
				err = podcast.HandleHTTP(w, r)
				if err != nil {
					s.handleError(w, r, err)
				}
				return
			}
		}

		// 403 would already be written if the user doesn't have access
		// therefore we aren't leaking any data
		w.WriteHeader(404)
	}
}

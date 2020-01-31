package main

import (
	"net/http"
)

func (s *server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("podcast_session")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/auth", 302)
				return
			}
			s.handleError(w, r, err)
			return
		}

		session := c.Value
		w.Write([]byte(session))
	}
}

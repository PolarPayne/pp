package main

import (
	"fmt"
	"net/http"
	"net/url"
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

		q := url.Values{}
		q.Set("s", c.Value)

		url := s.baseURL + "/feed?" + q.Encode()

		fmt.Fprintf(w, "Your personal RSS feed URL is %v\nAccesses to this URL are tracked, DO NOT SHARE THE URL WITH ANYONE", url)
	}
}

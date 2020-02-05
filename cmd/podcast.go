package main

import (
	"log"
	"sort"
	"time"

	"github.com/polarpayne/pp"
)

type podcastList []pp.Podcast

func (a podcastList) Len() int      { return len(a) }
func (a podcastList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a podcastList) Less(i, j int) bool {
	lhs, rhs := a[i].Details(), a[j].Details()

	if lhs.Published.Equal(rhs.Published) {
		return lhs.Title < rhs.Title
	}
	return rhs.Published.Before(lhs.Published)
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
		pd := p.Details()
		if now.Before(pd.Published) {
			log.Printf("updating podcasts: skipping podcast with published date in the future title=%q published=%v", pd.Title, pd.Published)
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

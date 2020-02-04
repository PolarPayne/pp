package pp

import (
	"net/http"
	"time"
)

type Podcast interface {
	Details() PodcastDetails
	HandlePodcast(http.ResponseWriter, *http.Request) error
}

type PodcastDetails struct {
	Key       string
	Title     string
	Published time.Time
	Size      int64
}

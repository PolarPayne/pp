package pp

import (
	"io"
	"time"
)

type Podcast interface {
	ID() string
	Title() string
	Published() time.Time
	Size() int64
	Open() (io.ReadCloser, error)
}

type Backend interface {
	ListPodcasts() ([]Podcast, error)
	GetPodcast(string) (Podcast, error)
}

package pp

import "io"

type Backend interface {
	GetLogo() (io.ReadCloser, error)
	ListPodcasts() ([]Podcast, error)
	GetPodcast(key string) (Podcast, error)
}

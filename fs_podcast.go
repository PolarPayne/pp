package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type FSPodcast struct {
	path  string
	title string
}

func NewFSPodcast(p string) FSPodcast {
	out := FSPodcast{}

	out.path = p
	out.title = strings.Split(path.Base(p), ".")[0]

	return out
}

func (p FSPodcast) Title() string {
	return p.title
}

func (p FSPodcast) Open() (io.Reader, error) {
	return os.Open(p.path)
}

func NewFSPodcasts(dir string) (PodcastsFunc, error) {
	s, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, fmt.Errorf("not a directory: %v", dir)
	}

	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]Podcast, 0, len(fs))
	for _, f := range fs {
		if f.IsDir() {
			continue
		}

		fullPath := path.Join(dir, f.Name())
		out = append(out, NewFSPodcast(fullPath))
	}

	f := func() ([]Podcast, error) {
		return out, nil
	}

	return f, nil
}

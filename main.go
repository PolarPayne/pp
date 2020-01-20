package main

import (
	"fmt"
	"io"

	"github.com/eduncan911/podcast"
)

type Podcast interface {
	Title() string
	Open() (io.Reader, error)
}

type PodcastsFunc func() ([]Podcast, error)

func main() {
	f, err := NewFSPodcasts("./podcasts")
	if err != nil {
		panic(err)
	}

	out := podcast.New("Reaktor", "http://localhost:8080/", "Reaktor Podcast", nil, nil)

	ps, _ := f()
	bs := make([]byte, 8)

	for _, p := range ps {
		fmt.Println(p)

		r, err := p.Open()
		if err != nil {
			panic(err)
		}

		n, err := r.Read(bs)
		if err != nil {
			panic(err)
		}
		if n != 8 {
			panic("couldn't read 8 bytes")
		}

		fmt.Println(bs)

		out.AddItem(podcast.Item{
			Title:       p.Title(),
			Description: "",
			Enclosure: podcast.Enclosure{

			}
		})
	}
}

package pp

import (
	"errors"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func splitTitle(name string) (time.Time, string, error) {
	var (
		t     time.Time
		title string
		err   error
	)

	split := strings.SplitN(name, " ", 2)
	if len(split) != 2 {
		return t, title, errors.New("")
	}

	t, err = time.Parse("2006-01-02", split[0])
	if err != nil {
		return t, title, errors.New("invalid date format (expected YYYY-MM-DD)")
	}

	title = split[1]
	if !strings.HasSuffix(title, ".mp3") {
		return t, title, errors.New("invalid title to end with .mp3")
	}

	return t, title[:len(title)-len(".mp3")], nil
}

type Podcast struct {
	backend   *Backend
	Key       string
	Size      int64
	Published time.Time
	Title     string
}

func newPodcast(backend *Backend, key string, size *int64) (Podcast, error) {
	if size == nil {
		return Podcast{}, errors.New("invalid Podcast: size is nil")
	}

	published, title, err := splitTitle(key)
	if err != nil {
		return Podcast{}, err
	}

	return Podcast{backend, key, *size, published, title}, nil
}

func (p *Podcast) Open() (io.ReadCloser, error) {
	obj, err := p.backend.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(p.backend.bucket),
		Key:    aws.String(p.Key),
	})
	if err != nil {
		return nil, err
	}

	return obj.Body, nil
}
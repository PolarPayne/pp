package pp

import (
	"errors"
	"io"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type BackendS3 struct {
	s3     *s3.S3
	bucket string
}

func NewBackendS3(sess *session.Session, bucket string) Backend {
	out := new(BackendS3)
	out.s3 = s3.New(sess)
	out.bucket = bucket

	return out
}

func (b *BackendS3) ListPodcasts() ([]Podcast, error) {
	r, err := b.s3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(b.bucket),
	})
	if err != nil {
		return nil, err
	}

	out := make([]Podcast, 0, len(r.Contents))
	for _, obj := range r.Contents {
		if obj.Key == nil {
			return nil, errors.New("invalid S3 object: Key or Size is nil")
		}

		key := *obj.Key
		if !strings.HasSuffix(key, ".mp3") {
			log.Printf("skipping non-MP3 file: %v", key)
			continue
		}

		p, err := newPodcastS3(b, key, obj.Size)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	return out, nil
}

func (b *BackendS3) GetPodcast(name string) (Podcast, error) {
	p, err := b.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	return newPodcastS3(b, name, p.ContentLength)
}

type PodcastS3 struct {
	backend   *BackendS3
	key       string
	size      int64
	title     string
	published time.Time
}

func newPodcastS3(backend *BackendS3, key string, size *int64) (*PodcastS3, error) {
	if size == nil {
		return nil, errors.New("invalid PodcastS3: size is nil")
	}

	out := new(PodcastS3)
	out.backend = backend
	out.key = key
	out.size = *size
	published, title, err := splitTitle(key)
	if err != nil {
		return nil, err
	}
	out.published = published
	out.title = title

	return out, nil
}

func (p *PodcastS3) ID() string {
	return p.key
}

func (p *PodcastS3) Title() string {
	return p.title
}

func (p *PodcastS3) Published() time.Time {
	return p.published
}

func (p *PodcastS3) Size() int64 {
	return p.size
}

func (p *PodcastS3) Open() (io.ReadCloser, error) {
	obj, err := p.backend.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(p.backend.bucket),
		Key:    aws.String(p.key),
	})
	if err != nil {
		return nil, err
	}

	return obj.Body, nil
}

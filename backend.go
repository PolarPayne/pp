package pp

import (
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Backend struct {
	s3     *s3.S3
	bucket string
}

func NewBackend(sess *session.Session, bucket string) *Backend {
	out := new(Backend)
	out.s3 = s3.New(sess)
	out.bucket = bucket

	return out
}

func (b *Backend) ListPodcasts() ([]Podcast, error) {
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

		p, err := newPodcast(b, key, obj.Size)
		if err != nil {
			log.Printf("invalid podcast %+v: %v", p, err)
			continue
		}
		out = append(out, p)
	}

	return out, nil
}

func (b *Backend) GetPodcast(key string) (Podcast, error) {
	p, err := b.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return Podcast{}, err
	}

	return newPodcast(b, key, p.ContentLength)
}

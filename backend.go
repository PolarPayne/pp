package pp

import (
	"errors"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Backend struct {
	s3     *s3.S3
	bucket string
	logo   string
}

func NewBackend(sess *session.Session, bucket string, logo string) *Backend {
	out := new(Backend)
	out.s3 = s3.New(sess)
	out.bucket = bucket
	out.logo = logo

	return out
}

func (b *Backend) GetLogo() (io.ReadCloser, error) {
	p, err := b.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.logo),
	})
	if err != nil {
		return nil, err
	}

	return p.Body, nil
}

func (b *Backend) ListPodcasts() ([]Podcast, error) {
	r, err := b.s3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(b.bucket),
	})
	if err != nil {
		return nil, err
	}

	contents := r.Contents

	// the S3 API is kinda annoying and we have to jump through some hoops to
	// make sure we get all of the objects, the go API for this is also not particularly great

	var (
		isTruncated bool
		nextMarker  string
	)

	for {
		if r.IsTruncated != nil {
			isTruncated = *r.IsTruncated
		}
		if r.NextMarker != nil {
			nextMarker = *r.NextMarker
		} else {
			isTruncated = false
		}

		if !isTruncated {
			break
		}

		r, err = b.s3.ListObjects(&s3.ListObjectsInput{
			Marker: &nextMarker,
		})
		if err != nil {
			return nil, err
		}
		contents = append(contents, r.Contents...)
	}

	out := make([]Podcast, 0, len(contents))
	for _, obj := range contents {
		if obj.Key == nil {
			return nil, errors.New("invalid S3 object: Key is nil")
		}

		key := *obj.Key
		if !strings.HasSuffix(key, ".mp3") {
			log.Printf("skipping non-MP3 file: %v", key)
			continue
		}

		p, err := newPodcast(b, key, obj.Size)
		if err != nil {
			log.Printf("invalid podcast: %v", err)
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

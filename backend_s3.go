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

type BackendS3 struct {
	s3     *s3.S3
	bucket string
	logo   string
}

func NewBackendS3(bucket, logo string) BackendS3 {
	session := session.Must(session.NewSession())
	return BackendS3{s3.New(session), bucket, logo}
}

func (b BackendS3) GetLogo() (io.ReadCloser, error) {
	p, err := b.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.logo),
	})
	if err != nil {
		return nil, err
	}

	return p.Body, nil
}

func (b BackendS3) ListPodcasts() ([]Podcast, error) {
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

		p, err := newPodcastS3(&b, key, obj.Size)
		if err != nil {
			log.Printf("invalid podcast: %v", err)
			continue
		}
		out = append(out, p)
	}

	return out, nil
}

func (b BackendS3) GetPodcast(key string) (Podcast, error) {
	p, err := b.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return PodcastS3{}, err
	}

	return newPodcastS3(&b, key, p.ContentLength)
}

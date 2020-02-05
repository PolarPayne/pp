package pp

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
		return t, title, fmt.Errorf("can't split name %q to date and title parts", name)
	}

	published := split[0]
	t, err = time.Parse("2006-01-02", published)
	if err != nil {
		return t, title, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %v", published)
	}

	title = split[1]
	if !strings.HasSuffix(title, ".mp3") {
		return t, title, fmt.Errorf("invalid title (expected it to end with .mp3): %v", title)
	}

	return t, title[:len(title)-len(".mp3")], nil
}

type PodcastS3 struct {
	backend     *BackendS3
	key         string
	size        int64
	published   time.Time
	title       string
	description string
}

func newPodcastS3(backend *BackendS3, key string, size *int64) (Podcast, error) {
	if size == nil {
		return PodcastS3{}, errors.New("size must be set: size is nil")
	}

	published, title, err := splitTitle(key)
	if err != nil {
		return PodcastS3{}, err
	}

	descriptionKey := key + ".txt"
	var description string
	obj, err := backend.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(backend.bucket),
		Key:    aws.String(descriptionKey),
	})
	if err == nil {
		defer obj.Body.Close()
		desc, err := ioutil.ReadAll(obj.Body)
		if err == nil {
			description = string(desc)
		} else {
			log.Printf("failed to read description of PodcastS3 key=%q: %v", descriptionKey, err)
		}
	} else {
		log.Printf("failed to get description of PodcastS3 key=%q: %v", descriptionKey, err)
	}

	return PodcastS3{backend, key, *size, published, title, description}, nil
}

func (p PodcastS3) Details() PodcastDetails {
	return PodcastDetails{
		Key:         p.key,
		Title:       p.title,
		Published:   p.published,
		Size:        p.size,
		Description: p.description,
	}
}

func (p PodcastS3) HandlePodcast(w http.ResponseWriter, r *http.Request) error {
	w.Header().Add("Content-Type", "audio/mpeg")
	w.Header().Add("Accept-Ranges", "bytes")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		return p.handleRangeHeader(w, r, rangeHeader)
	}

	return p.handleNormal(w, r)
}

func (p PodcastS3) handleRangeHeader(w http.ResponseWriter, r *http.Request, rangeHeader string) error {
	log.Printf("request with Range header: %v", rangeHeader)

	obj, err := p.backend.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(p.backend.bucket),
		Key:    aws.String(p.key),
		Range:  aws.String(rangeHeader),
	})
	if err != nil {
		return fmt.Errorf("failed to get object from S3 w/ Range header: %v", err)
	}
	defer obj.Body.Close()

	contentLength, contentRange := obj.ContentLength, obj.ContentRange
	if contentLength == nil || contentRange == nil {
		return fmt.Errorf("S3 returned nil Content-Length (=%q) and/or Content-Range (=%q)", contentLength, contentRange)
	}

	w.Header().Set("Content-Length", strconv.FormatInt(*contentLength, 10))
	w.Header().Set("Content-Range", *contentRange)

	w.WriteHeader(http.StatusPartialContent)
	n, err := io.Copy(w, obj.Body)
	if err != nil {
		return fmt.Errorf("failed to copy %v bytes of content to response: %v", n, err)
	}

	return nil
}

func (p PodcastS3) handleNormal(w http.ResponseWriter, r *http.Request) error {
	obj, err := p.backend.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(p.backend.bucket),
		Key:    aws.String(p.key),
	})
	if err != nil {
		return fmt.Errorf("failed get object from S3: %v", err)
	}

	defer obj.Body.Close()

	contentLength := obj.ContentLength
	if contentLength == nil {
		return fmt.Errorf("S3 returned nil Content-Length")
	}

	w.Header().Add("Content-Length", strconv.FormatInt(*contentLength, 10))

	n, err := io.Copy(w, obj.Body)
	if err != nil {
		return fmt.Errorf("failed to copy content to response (%v of %v bytes copied): %v", n, *contentLength, err)
	}

	return nil
}

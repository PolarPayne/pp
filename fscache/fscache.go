package fscache

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func hashToString(hash hash.Hash) string {
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

var ErrNotExists = errors.New("fscache: file with such key does not exist in cache")

type FSCache struct {
	tmpDir string
}

func New() (FSCache, error) {
	dir, err := ioutil.TempDir("", "pp-*")
	if err != nil {
		return FSCache{}, err
	}

	return FSCache{dir}, nil
}

func (fc FSCache) Close() error {
	return os.RemoveAll(fc.tmpDir)
}

func (fc FSCache) Set(r io.Reader) (string, error) {
	tmpFile, err := ioutil.TempFile("", "pp-put-tmp-*")
	if err != nil {
		return "", err
	}

	cleanTmpFile := func(originalError error) error {
		if err = os.Remove(tmpFile.Name()); err != nil {
			return err
		}
		return originalError
	}

	hash := sha256.New()
	w := io.MultiWriter(hash, tmpFile)

	_, err = io.Copy(w, r)
	if err != nil {
		return "", cleanTmpFile(err)
	}

	key := hashToString(hash)
	cachePath := path.Join(fc.tmpDir, key)
	// a file with the exact same hash is already in this FSCache
	if _, err = os.Stat(cachePath); !os.IsNotExist(err) {
		// just cleanup the created tmp file and return the key
		cleanTmpFile(nil)
		return key, nil
	}

	err = os.Rename(tmpFile.Name(), cachePath)
	if err != nil {
		return "", err
	}

	return key, nil
}

func (fc FSCache) Get(key string) (io.ReadCloser, error) {
	path := path.Join(fc.tmpDir, key)
	fp, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotExists
		}

		return nil, err
	}

	return fp, nil
}

func (fc FSCache) GetPath(key string) (string, error) {
	path := path.Join(fc.tmpDir, key)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotExists
		}

		return "", err
	}

	return path, nil
}

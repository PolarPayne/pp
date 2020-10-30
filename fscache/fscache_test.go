package fscache_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/polarpayne/pp/fscache"
	"github.com/stretchr/testify/assert"
)

func TestBasicUsage(t *testing.T) {
	assert := assert.New(t)

	// New

	fc, err := fscache.New()
	assert.NoError(err)

	// Set

	data := bytes.Buffer{}
	data.WriteString("Hello World!")

	key, err := fc.Set(&data)
	assert.NoError(err)

	// Get

	r, err := fc.Get(key)
	assert.NoError(err)

	returnData, err := ioutil.ReadAll(r)
	assert.NoError(err)

	assert.Equal("Hello World!", string(returnData))

	// GetPath

	path, err := fc.GetPath(key)
	assert.NoError(err)

	returnDataPath, err := ioutil.ReadFile(path)
	assert.NoError(err)

	assert.Equal("Hello World!", string(returnDataPath))

	// Close

	err = fc.Close()
	assert.NoError(err)
}

func TestGetNonExistent(t *testing.T) {
	assert := assert.New(t)

	fc, err := fscache.New()
	assert.NoError(err)

	r, err := fc.Get("nonexistent")
	assert.Nil(r)
	assert.Error(err)
	assert.Equal(fscache.ErrNotExists, err)
}

func TestSetSameContent(t *testing.T) {
	assert := assert.New(t)

	fc, err := fscache.New()
	assert.NoError(err)

	data1 := bytes.Buffer{}
	data1.WriteString("Hello World!")

	key1, err := fc.Set(&data1)
	assert.NoError(err)

	data2 := bytes.Buffer{}
	data2.WriteString("Hello World!")

	key2, err := fc.Set(&data2)
	assert.NoError(err)

	assert.Equal(key1, key2)
}

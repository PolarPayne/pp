package pp

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

type Storage interface {
	Init() error
	CreateUser(userID string) (string, error)
	ValidSecret(secret string) (bool, error)
	// LogAudit(secret, referer, userAgent string) error
}

// SecretSizeBytes is the size of the secret in bytes, it should be a multiple of 12 to make sure it's encoded nicely in base64.
const SecretSizeBytes = 36

// GenerateSecret generates `SecretSizeBytes` bytes of secure randomness and encodes it with URL safe base64.
func GenerateSecret() string {
	v := make([]byte, SecretSizeBytes)
	n, err := rand.Read(v)
	if err != nil || n != SecretSizeBytes {
		log.Panicf("failed to create random %v bytes: %v", SecretSizeBytes, err)
	}

	return base64.URLEncoding.EncodeToString(v)
}

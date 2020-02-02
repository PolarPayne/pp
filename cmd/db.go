package main

import "math/rand"

import "fmt"

import "encoding/base64"

import "log"

const secretSize = 36

func (s *server) createUser(userID string) (string, error) {
	var secret string

	// if a user with the given userID already exists return that users secret
	// instead, this makes sure the user can access the home page even if their
	// cookies are reseted (or if they use multiple machines)
	err := s.db.QueryRow(`SELECT secret FROM users WHERE user_id = $1`, userID).Scan(&secret)
	if err == nil {
		return secret, nil
	}

	v := make([]byte, secretSize)
	n, err := rand.Read(v)
	if err != nil {
		return "", err
	}
	if n != secretSize {
		return "", fmt.Errorf("failed to read %v random bytes, only %v read", secretSize, n)
	}

	secret = base64.URLEncoding.EncodeToString(v)

	_, err = s.db.Exec(`INSERT INTO users (user_id, secret) VALUES ($1, $2)`, userID, secret)
	if err != nil {
		return "", fmt.Errorf("failed to insert user into db: %v", err)
	}

	return secret, nil
}

func (s *server) validSecret(secret string) (bool, error) {
	var userID string

	err := s.db.QueryRow(`SELECT user_id FROM users WHERE secret = $1`, secret).Scan(&userID)
	if err != nil {
		return false, err
	}

	log.Printf("fetched the secret of user %q from the db", userID)
	return true, nil
}

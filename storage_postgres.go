package pp

import (
	"database/sql"
	"fmt"
	"log"

	// pq is used through database/sql by StoragePostgres
	_ "github.com/lib/pq"
)

type StoragePostgres struct {
	db *sql.DB
}

func NewStoragePostgres(connectionString string) (StoragePostgres, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return StoragePostgres{}, err
	}

	return StoragePostgres{db}, nil
}

func (s StoragePostgres) Init() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		user_id    TEXT UNIQUE NOT NULL,
		secret     TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`)

	if err == nil {
		_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS log_feed (
			secret     TEXT NOT NULL,
			referer    TEXT NOT NULL,
			user_agent TEXT NOT NULL,
			timestamp  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`)
	}

	if err == nil {
		_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS log_podcast (
			secret     TEXT NOT NULL,
			key        TEXT NOT NULL,
			referer    TEXT NOT NULL,
			user_agent TEXT NOT NULL,
			timestamp  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`)
	}

	if err != nil {
		return fmt.Errorf("failed to create table(s) in db: %v", err)
	}

	log.Print("created tables in the database")
	return nil
}

func (s StoragePostgres) CreateUser(userID string) (string, error) {
	var secret string

	// if a user with the given userID already exists return that users secret
	// instead, this makes sure the user can access the home page even if their
	// cookies are reseted (or if they use multiple machines)
	err := s.db.QueryRow(`SELECT secret FROM users WHERE user_id = $1`, userID).Scan(&secret)
	if err == nil {
		return secret, nil
	}

	secret = GenerateSecret()

	_, err = s.db.Exec(`INSERT INTO users (user_id, secret) VALUES ($1, $2)`, userID, secret)
	if err != nil {
		return "", fmt.Errorf("failed to insert user into db: %v", err)
	}

	return secret, nil
}

func (s StoragePostgres) ValidSecret(secret string) (bool, error) {
	var userID string

	err := s.db.QueryRow(`SELECT user_id FROM users WHERE secret = $1`, secret).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("invalid secret %q", secret)
			return false, nil
		}
		return false, err
	}

	log.Printf("fetched the secret of user %q from the db", userID)
	return true, nil
}

func (s StoragePostgres) LogFeed(secret, referer, userAgent string) error {
	_, err := s.db.Exec(
		`INSERT INTO log_feed (secret, referer, user_agent)
		VALUES ($1, $2, $3)`,
		secret, referer, userAgent)
	if err != nil {
		return fmt.Errorf("failed to insert into db: %v", err)
	}

	return nil
}

func (s StoragePostgres) LogPodcast(secret, key, referer, userAgent string) error {
	_, err := s.db.Exec(
		`INSERT INTO log_podcast (secret, key, referer, user_agent)
		VALUES ($1, $2, $3, $4)`,
		secret, key, referer, userAgent)
	if err != nil {
		return fmt.Errorf("failed to insert into db: %v", err)
	}

	return nil
}

package pgsqldb

import (
	"database/sql"
	"encoding/hex"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/pkg/generator"
	"golang.org/x/crypto/argon2"
)

type userStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *userStorage {
	return &userStorage{db: db}
}

var salt = []byte("3a1tfor5a44word123")

func (s userStorage) Get(username, password string) (*domain.User, error) {
	user := &domain.User{}
	hashedPass := getHashPassword(password)
	row := s.db.QueryRow("SELECT user_id FROM users WHERE username=$1 AND password=$2 AND is_active=True", username, hashedPass)

	var userID int
	errScan := row.Scan(&userID)

	if errScan != nil {
		return nil, errScan
	}

	user.Username = username
	user.Password = password
	user.ID = generator.GenerateNewIDByMD(username)

	return user, nil
}

func (s userStorage) Save(username, password string) (*domain.User, error) {
	user := &domain.User{}
	user.Username = username
	user.Password = password

	hashedPass := getHashPassword(password)
	_, err := s.db.Exec(`INSERT INTO users (username, password, is_active) VALUES ($1, $2, $3)`, username, hashedPass, true)
	if err != nil {
		return nil, err
	}

	user.ID = generator.GenerateNewIDByMD(username)

	return user, nil
}

func getHashPassword(password string) string {
	pass := []byte(password)
	hashedPass := argon2.IDKey(pass, salt, 1, 64*1024, 4, 32)
	return hex.EncodeToString(hashedPass)
}

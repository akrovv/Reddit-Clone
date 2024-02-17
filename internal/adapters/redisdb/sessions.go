package redisdb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/dgrijalva/jwt-go"
	"github.com/redis/go-redis/v9"
)

var tokenSecret = []byte("super secret key")

type sessionStorage struct {
	ctx context.Context
	db  *redis.Client
}

func NewSessionStorage(ctx context.Context, db *redis.Client) *sessionStorage {
	return &sessionStorage{ctx: ctx, db: db}
}

func (s sessionStorage) Create(ID, username string) (*domain.Session, error) {
	userWithoutPassword := &domain.User{
		Username: username,
		ID:       ID,
	}

	sess, errSession := newSession(userWithoutPassword)

	if errSession != nil {
		return nil, errSession
	}

	errSet := set(s.ctx, s.db, sess.ID, sess.User)

	if errSet != nil {
		return nil, errSet
	}

	return sess, nil
}

func (s sessionStorage) Get(key string) (*domain.User, error) {
	user := &domain.User{}
	err := get(s.ctx, s.db, key, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func set(ctx context.Context, client *redis.Client, key string, value any) error {
	p, err := json.Marshal(value)

	if err != nil {
		return err
	}

	status := client.Set(ctx, key, p, time.Hour*9)

	if status.Err() != nil {
		return status.Err()
	}

	return nil
}

func get(ctx context.Context, client *redis.Client, key string, dest any) error {
	value, err := client.Get(ctx, key).Result()

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(value), dest)
}

func newSession(user *domain.User) (*domain.Session, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
	})

	tokenString, err := token.SignedString(tokenSecret)
	if err != nil {
		return nil, err
	}

	return &domain.Session{
		ID:   tokenString,
		User: user,
	}, nil
}

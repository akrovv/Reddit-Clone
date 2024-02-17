package redisdb

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/go-redis/redismock/v9"
)

func TestCreate(t *testing.T) {
	ctx := context.TODO()
	db, mock := redismock.NewClientMock()
	repo := NewSessionStorage(ctx, db)
	key := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoiYWtyb3YiLCJwYXNzd29yZCI6IiIsImlkIjoiMSJ9fQ.zXYAcNhgO2NeKvqb0mBTMv_EGXWj2Y0ZQYnREesJGM4"
	userWithoutPassword := &domain.User{
		Username: "akrov",
		ID:       "1",
	}

	sessionExpected := &domain.Session{
		ID:   key,
		User: userWithoutPassword,
	}

	p, err := json.Marshal(userWithoutPassword)

	if err != nil {
		t.Error("test error")
		return
	}

	// OK
	mock.ExpectSet(key,
		p,
		time.Hour*9).SetVal(key)

	session, err := repo.Create(userWithoutPassword.ID, userWithoutPassword.Username)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if !reflect.DeepEqual(sessionExpected, session) {
		t.Errorf("expected: %v, got: %v", sessionExpected, session)
		return
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	// Set error
	mock.ExpectSet(key,
		p,
		time.Hour*9).SetErr(errors.New("some error"))

	session, err = repo.Create(userWithoutPassword.ID, userWithoutPassword.Username)

	if err == nil {
		t.Error("expected error, got nil")
		return
	}

	if session != nil {
		t.Error("expected nil, got session")
		return
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}

func TestGet(t *testing.T) {
	ctx := context.TODO()
	db, mock := redismock.NewClientMock()
	repo := NewSessionStorage(ctx, db)
	userExpected := &domain.User{
		Username: "akro",
		ID:       "1",
	}
	key := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoiYWtyb3YiLCJwYXNzd29yZCI6IiIsImlkIjoiMSJ9fQ.zXYAcNhgO2NeKvqb0mBTMv_EGXWj2Y0ZQYnREesJGM4"

	// OK
	mock.ExpectGet(key).SetVal(`{"username":"akro", "id":"1"}`)
	user, err := repo.Get(key)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if !reflect.DeepEqual(userExpected, user) {
		t.Errorf("expected: %v, got: %v", userExpected, user)
		return
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	// Get error
	mock.ExpectGet(key).SetErr(errors.New("some error"))
	user, err = repo.Get(key)

	if err == nil {
		t.Error("expected error, got nil")
		return
	}

	if user != nil {
		t.Error("expected nil, got user")
		return
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}

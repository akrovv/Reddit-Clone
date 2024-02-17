package pgsqldb

import (
	"errors"
	"reflect"
	"testing"

	"github.com/akrovv/redditclone/internal/domain"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestSave(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("can't create mock: %s", err)
		return
	}
	defer db.Close()

	repo := NewUserStorage(db)

	username := "username"
	password := "password"
	isActive := true
	hashedPass := getHashPassword(password)

	userExpected := &domain.User{
		Username: username,
		Password: password,
		ID:       "14c4b06b824ec59323936251",
	}

	// OK
	mock.ExpectExec(`INSERT INTO users`).WithArgs(username, hashedPass, isActive).WillReturnResult(sqlmock.NewResult(1, 1))
	user, err := repo.Save(username, password)

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
	}

	// Save error
	mock.ExpectExec(`INSERT INTO users`).WithArgs(username, hashedPass, isActive).WillReturnError(errors.New("some error"))

	user, err = repo.Save(username, password)

	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	if user != nil {
		t.Error("expected nil, got user")
		return
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGet(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("can't create mock: %s", err)
		return
	}
	defer db.Close()

	repo := NewUserStorage(db)

	userID := 1
	username := "username"
	password := "password"
	hashedPass := getHashPassword(password)

	userExpected := &domain.User{
		Username: username,
		Password: password,
		ID:       "14c4b06b824ec59323936251",
	}

	// OK
	rows := sqlmock.NewRows([]string{"user_id"}).AddRow(userID)
	mock.ExpectQuery("SELECT user_id FROM users WHERE").WithArgs(username, hashedPass).WillReturnRows(rows)
	user, err := repo.Get(username, password)

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
	}

	mock.ExpectQuery("SELECT user_id FROM users WHERE").WithArgs(username, hashedPass).WillReturnError(errors.New("some error"))
	user, err = repo.Get(username, password)

	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	if user != nil {
		t.Error("expected nil, got user")
		return
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

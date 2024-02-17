package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"strings"
	"testing"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/internal/service"
	"github.com/akrovv/redditclone/internal/service/mocks"
	"github.com/akrovv/redditclone/pkg/logger"
	"github.com/golang/mock/gomock"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mocks.NewMockUserService(ctrl)
	ss := mocks.NewMockSessionService(ctrl)

	userHandler := NewUserHandler(logger.NewLogger(), us, ss)

	body := `{
		"username": "akro",
		  "password": "akroakroakro"
	  }`

	user := &domain.User{}
	session := &domain.Session{}

	userGetDto := &service.GetUser{Username: "akro"}
	userSaveDto := &service.SaveUser{Username: "akro", Password: "akroakroakro"}
	sessionCreateDto := &service.CreateSession{}

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(body))

	req.Header.Add("Content-type", "application/json")
	w := httptest.NewRecorder()

	// Correct answer
	us.EXPECT().Get(userGetDto).Return(nil, nil)
	us.EXPECT().Save(userSaveDto).Return(user, nil)
	ss.EXPECT().Create(sessionCreateDto).Return(session, nil)

	userHandler.Register(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Bad header
	req.Header.Del("Content-type")
	w = httptest.NewRecorder()

	userHandler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got: %d", w.Code)
		return
	}

	// Bad Unmarshall
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(`{"username": {`))
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	userHandler.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad Check
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(body))
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	us.EXPECT().Get(userGetDto).Return(nil, errors.New("some error"))

	userHandler.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad user, already exists
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(body))
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()
	us.EXPECT().Get(userGetDto).Return(user, nil)
	userHandler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad Save
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(body))
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	us.EXPECT().Get(userGetDto).Return(nil, nil)
	us.EXPECT().Save(userSaveDto).Return(nil, errors.New("some error"))

	userHandler.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad session Create
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(body))
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	us.EXPECT().Get(userGetDto).Return(nil, nil)
	us.EXPECT().Save(userSaveDto).Return(user, nil)
	ss.EXPECT().Create(sessionCreateDto).Return(nil, errors.New("some error"))

	userHandler.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad readAll
	req = httptest.NewRequest("POST", "/api/register", &BadReader{})
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	userHandler.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad write
	req = httptest.NewRequest("POST", "/api/register", strings.NewReader(body))

	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	us.EXPECT().Get(userGetDto).Return(nil, nil)
	us.EXPECT().Save(userSaveDto).Return(user, nil)
	ss.EXPECT().Create(sessionCreateDto).Return(session, nil)

	userHandler.Register(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mocks.NewMockUserService(ctrl)
	ss := mocks.NewMockSessionService(ctrl)

	userHandler := NewUserHandler(logger.NewLogger(), us, ss)

	body := `{
		"username": "akro",
		  "password": "akroakroakro"
	  }`

	user := &domain.User{}
	session := &domain.Session{}

	userGetDto := &service.GetUser{Username: "akro", Password: "akroakroakro"}
	sessionCreateDto := &service.CreateSession{}

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(body))

	req.Header.Add("Content-type", "application/json")
	w := httptest.NewRecorder()

	// Correct answer
	us.EXPECT().Get(userGetDto).Return(user, nil)
	ss.EXPECT().Create(sessionCreateDto).Return(session, nil)

	userHandler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Bad header
	req.Header.Del("Content-type")
	w = httptest.NewRecorder()

	userHandler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got: %d", w.Code)
		return
	}

	// Bad Unmarshall
	req = httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"username": {`))
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	userHandler.Login(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad check
	req = httptest.NewRequest("POST", "/api/login", strings.NewReader(body))
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	us.EXPECT().Get(userGetDto).Return(nil, errors.New("some error"))

	userHandler.Login(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad session Create
	req = httptest.NewRequest("POST", "/api/login", strings.NewReader(body))
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	us.EXPECT().Get(userGetDto).Return(user, nil)
	ss.EXPECT().Create(sessionCreateDto).Return(nil, errors.New("some error"))

	userHandler.Login(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad readAll
	req = httptest.NewRequest("POST", "/api/login", &BadReader{})
	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	userHandler.Login(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Bad Write
	req = httptest.NewRequest("POST", "/api/login", strings.NewReader(body))

	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	us.EXPECT().Get(userGetDto).Return(user, nil)
	ss.EXPECT().Create(sessionCreateDto).Return(session, nil)

	userHandler.Login(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

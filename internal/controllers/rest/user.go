package rest

import (
	"encoding/json"

	"io"
	"net/http"

	"github.com/akrovv/redditclone/internal/service"
	jsontransfer "github.com/akrovv/redditclone/pkg/jsonTransfer"
	"github.com/akrovv/redditclone/pkg/logger"
)

type userHandler struct {
	logger         logger.Logger
	userService    UserService
	sessionService SessionService
}

func NewUserHandler(logger logger.Logger, userService UserService, sessionService SessionService) *userHandler {
	return &userHandler{logger: logger, userService: userService, sessionService: sessionService}
}

func (h userHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-type") != "application/json" {
		h.logger.Infof("not found application/json header")
		http.Error(w, "not found application/json header", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)

	if err != nil {
		h.logger.Infof("can't read form: %w", err)
		http.Error(w, "can't read form", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	getUserDto := &service.GetUser{}
	err = json.Unmarshal(data, getUserDto)

	if err != nil {
		h.logger.Infof("can't unmarshall json to getUser: %w", err)
		http.Error(w, "can't unmarshall json", http.StatusInternalServerError)
		return
	}

	user, err := h.userService.Get(getUserDto)

	if err != nil {
		h.logger.Infof("can't get user: %w", err)
		http.Error(w, "can't get user", http.StatusInternalServerError)
		return
	}

	sessionDto := &service.CreateSession{
		ID:       user.ID,
		Username: user.Username,
	}

	sess, err := h.sessionService.Create(sessionDto)

	if err != nil {
		h.logger.Infof("can't create session: %w", err)
		http.Error(w, "can't create session", http.StatusInternalServerError)
		return
	}

	token, err := jsontransfer.GetJSON(struct {
		Token string `json:"token"`
	}{
		Token: sess.ID,
	})

	if err != nil {
		h.logger.Infof("can't marshall token to json: %w", err)
		http.Error(w, "can't marhsall token to json", http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[%s] %s created session for user  with session's id [%s]", r.Method, r.URL.Path, sess.ID)
	w.Header().Add("Content-type", "application/json")

	_, err = w.Write(token)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h userHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-type") != "application/json" {
		h.logger.Infof("not found application/json header")
		http.Error(w, "not found application/json header", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Infof("can't read form: %w", err)
		http.Error(w, "can't read form", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	userSaveDto := &service.SaveUser{}
	err = json.Unmarshal(data, userSaveDto)

	if err != nil {
		h.logger.Infof("can't unmarshall json to getUser: %w", err)
		http.Error(w, "can't unmarshall json", http.StatusInternalServerError)
		return
	}

	userGetDto := &service.GetUser{Username: userSaveDto.Username}
	user, err := h.userService.Get(userGetDto)

	if err != nil && err.Error() != "sql: no rows in result set" {
		h.logger.Infof("can't checked user: %w", err)
		http.Error(w, "can't checked user", http.StatusInternalServerError)
		return
	}

	if user != nil {
		h.logger.Info("user already exists")
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}

	user, err = h.userService.Save(userSaveDto)

	if err != nil {
		h.logger.Infof("can't save user in repo: %w", err)
		http.Error(w, "can't save user in repo", http.StatusInternalServerError)
		return
	}

	sessionDto := &service.CreateSession{
		ID:       user.ID,
		Username: user.Username,
	}

	sess, err := h.sessionService.Create(sessionDto)

	if err != nil {
		h.logger.Infof("can't create session: %w", err)
		http.Error(w, "can't create session", http.StatusInternalServerError)
		return
	}

	token, err := jsontransfer.GetJSON(struct {
		Token string `json:"token"`
	}{
		Token: sess.ID,
	})

	if err != nil {
		h.logger.Infof("can't marshall token to json: %w", err)
		http.Error(w, "can't marhsall token to json", http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[%s] %s created session for user with session's id [%s]", r.Method, r.URL.Path, sess.ID)
	w.Header().Add("Content-type", "application/json")
	_, err = w.Write(token)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

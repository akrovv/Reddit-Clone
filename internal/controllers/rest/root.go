package rest

import (
	"net/http"

	"github.com/akrovv/redditclone/pkg/logger"
)

type RootHandler struct {
	logger logger.Logger
}

func NewRootHandler(logger logger.Logger) *RootHandler {
	return &RootHandler{logger: logger}
}

func (h RootHandler) Main(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "front/html/index.html")
	h.logger.Info("sent file with static")
}

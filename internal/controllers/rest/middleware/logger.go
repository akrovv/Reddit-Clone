package middleware

import (
	"net/http"

	"time"

	"github.com/akrovv/redditclone/pkg/logger"
)

func Logger(next http.Handler, logger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		next.ServeHTTP(w, r)
		logger.Infof("[%s] %s timeAnswer=%v", r.Method, r.URL.Path, time.Since(t))
	})
}

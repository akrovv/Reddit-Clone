package middleware

import (
	"net/http"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/pkg/logger"
	"github.com/casbin/casbin"
)

func Permissions(next http.Handler, logger logger.Logger, e *casbin.Enforcer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := "anonymous"
		ctx := r.Context()

		var sess domain.SessionContextKey = "session"
		sessCtx := ctx.Value(sess)

		if sessCtx != nil {
			role = "member"
		}

		res, err := e.EnforceSafe(role, r.URL.Path, r.Method)

		if err != nil {
			http.Error(w, "problem with request", http.StatusInternalServerError)
			return
		}

		logger.Infof("path=%s role=%s access=%v", r.URL.Path, role, res)
		if res {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	})
}

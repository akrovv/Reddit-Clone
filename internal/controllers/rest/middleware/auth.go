package middleware

import (
	"context"
	"net/http"

	"strings"

	"github.com/akrovv/redditclone/internal/controllers/rest"
	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/internal/service"
)

func Auth(next http.Handler, au rest.SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header != "" {
			var sess domain.SessionContextKey = "session"

			authorization := strings.TrimPrefix(header, "Bearer ")
			getSessionDto := &service.GetSession{Key: authorization}
			user, err := au.Get(getSessionDto)

			if err != nil {
				http.Redirect(w, r, "/", http.StatusInternalServerError)
				return
			}

			if user == nil && err == nil {
				http.Redirect(w, r, "/", http.StatusNotFound)
				return
			}

			ctx := r.Context()

			ctx = context.WithValue(ctx, sess, &domain.Session{User: user})

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r)
	})
}

package auth

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	dbsqlc "github.com/donamo/artline-backend/internal/db"
)

type contextKey string

const userContextKey contextKey = "user"

func SessionMiddleware(db *sql.DB, store *sessions.CookieStore) func(http.Handler) http.Handler {
	queries := dbsqlc.New(db)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := store.Get(r, sessionName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			userIDStr, ok := sess.Values[sessionUserKey].(string)
			if !ok || userIDStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := queries.GetUserByID(r.Context(), userID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserFromContext(ctx context.Context) *dbsqlc.User {
	u, _ := ctx.Value(userContextKey).(*dbsqlc.User)
	return u
}

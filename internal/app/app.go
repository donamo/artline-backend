package app

import (
	"database/sql"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"

	"github.com/donamo/artline-backend/internal/auth"
	"github.com/donamo/artline-backend/internal/graph"
)

type App struct {
	DB           *sql.DB
	SessionStore *sessions.CookieStore
	Router       http.Handler
}

func New(db *sql.DB, sessionSecret string) (*App, error) {
	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   false, // true lesz production-ban HTTPS mögött
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
		Path:     "/",
	}

	authHandler, err := auth.NewHandler(db, store)
	if err != nil {
		return nil, err
	}

	a := &App{
		DB:           db,
		SessionStore: store,
	}
	a.Router = a.buildRouter(authHandler, store)
	return a, nil
}

func (a *App) buildRouter(authHandler *auth.Handler, store *sessions.CookieStore) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(auth.SessionMiddleware(a.DB, store))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Get("/auth/google/login", authHandler.Login)
	r.Get("/auth/google/callback", authHandler.Callback)
	r.Post("/auth/logout", authHandler.Logout)

	r.With(auth.RequireAuth).Get("/auth/me", authHandler.Me)

	gqlSrv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: graph.NewResolver(a.DB),
	}))
	r.Handle("/graphql", gqlSrv)
	r.Handle("/playground", playground.Handler("GraphQL", "/graphql"))

	return r
}

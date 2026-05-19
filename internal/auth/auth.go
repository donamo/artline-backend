package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	dbsqlc "github.com/donamo/artline-backend/internal/db"
)

const sessionName = "artline_session"
const sessionUserKey = "user_id"

type Handler struct {
	store    *sessions.CookieStore
	queries  *dbsqlc.Queries
	oauth2   *oauth2.Config
	verifier *oidc.IDTokenVerifier
	frontendURL string
}

func NewHandler(db *sql.DB, store *sessions.CookieStore) (*Handler, error) {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		return nil, err
	}

	cfg := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Endpoint:     google.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return &Handler{
		store:       store,
		queries:     dbsqlc.New(db),
		oauth2:      cfg,
		verifier:    provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		frontendURL: frontendURL,
	}, nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	sess, _ := h.store.Get(r, sessionName)
	sess.Values["oauth_state"] = state
	sess.Save(r, w)

	http.Redirect(w, r, h.oauth2.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	sess, _ := h.store.Get(r, sessionName)

	savedState, _ := sess.Values["oauth_state"].(string)
	if savedState == "" || savedState != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	delete(sess.Values, "oauth_state")

	token, err := h.oauth2.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "missing id_token", http.StatusInternalServerError)
		return
	}

	idToken, err := h.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "invalid id_token", http.StatusInternalServerError)
		return
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "claims error", http.StatusInternalServerError)
		return
	}

	user, err := h.queries.UpsertUser(r.Context(), dbsqlc.UpsertUserParams{
		GoogleSubject: claims.Sub,
		Email:         claims.Email,
		DisplayName:   sql.NullString{String: claims.Name, Valid: claims.Name != ""},
	})
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	sess.Values[sessionUserKey] = user.ID.String()
	sess.Save(r, w)

	http.Redirect(w, r, h.frontendURL, http.StatusTemporaryRedirect)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := h.store.Get(r, sessionName)
	sess.Options.MaxAge = -1
	sess.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":          user.ID,
		"email":       user.Email,
		"displayName": user.DisplayName,
	})
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

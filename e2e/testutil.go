package e2e

import (
	"database/sql"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/donamo/artline-backend/db/migrations"
	"github.com/donamo/artline-backend/internal/app"
	dbsqlc "github.com/donamo/artline-backend/internal/db"
)

const testSessionSecret = "test-secret-do-not-use-in-production"

type Suite struct {
	DB      *sql.DB
	Server  *httptest.Server
	store   *sessions.CookieStore
	queries *dbsqlc.Queries
}

func NewSuite(t *testing.T) *Suite {
	t.Helper()

	_ = godotenv.Load("../.env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping test db: %v", err)
	}

	goose.SetBaseFS(migrations.FS)
	_ = goose.SetDialect("postgres")
	if err := goose.Up(db, "."); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	a, err := app.New(db, testSessionSecret)
	if err != nil {
		t.Fatalf("init app: %v", err)
	}

	srv := httptest.NewServer(a.Router)

	s := &Suite{
		DB:      db,
		Server:  srv,
		store:   a.SessionStore,
		queries: dbsqlc.New(db),
	}

	t.Cleanup(func() {
		s.truncateTables(t)
		srv.Close()
		db.Close()
	})

	return s
}

// LoginAs creates a test user and returns an http.Client with a valid session cookie.
func (s *Suite) LoginAs(t *testing.T, googleSubject, email string) (*http.Client, *dbsqlc.User) {
	t.Helper()

	user, err := s.queries.UpsertUser(t.Context(), dbsqlc.UpsertUserParams{
		GoogleSubject: googleSubject,
		Email:         email,
		DisplayName:   sql.NullString{String: email, Valid: true},
	})
	if err != nil {
		t.Fatalf("upsert test user: %v", err)
	}

	// encode a real session cookie using the same store as the app
	req, _ := http.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	sess, _ := s.store.Get(req, "artline_session")
	sess.Values["user_id"] = user.ID.String()
	sess.Save(req, rec)

	jar, _ := cookiejar.New(nil)
	serverURL, _ := url.Parse(s.Server.URL)
	jar.SetCookies(serverURL, rec.Result().Cookies())

	client := &http.Client{Jar: jar}
	return client, &user
}

func (s *Suite) truncateTables(t *testing.T) {
	t.Helper()
	_, err := s.DB.Exec(`truncate table creative_project_links, creative_projects, users restart identity cascade`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func (s *Suite) CreateTestUser(t *testing.T) *dbsqlc.User {
	t.Helper()
	sub := uuid.New().String()
	user, err := s.queries.UpsertUser(t.Context(), dbsqlc.UpsertUserParams{
		GoogleSubject: sub,
		Email:         sub + "@test.com",
		DisplayName:   sql.NullString{String: "Test User", Valid: true},
	})
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return &user
}

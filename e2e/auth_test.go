package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogout_NoSession(t *testing.T) {
	s := NewSuite(t)

	resp, err := http.Post(s.Server.URL+"/auth/logout", "", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogout_WithSession(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "google-sub-1", "user1@test.com")

	resp, err := client.Post(s.Server.URL+"/auth/logout", "", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMe_Unauthenticated(t *testing.T) {
	s := NewSuite(t)

	resp, err := http.Get(s.Server.URL + "/auth/me")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMe_Authenticated(t *testing.T) {
	s := NewSuite(t)
	client, user := s.LoginAs(t, "google-sub-2", "user2@test.com")

	resp, err := client.Get(s.Server.URL + "/auth/me")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "user2@test.com", user.Email)
}

func TestMe_AfterLogout(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "google-sub-3", "user3@test.com")

	resp, err := client.Post(s.Server.URL+"/auth/logout", "", nil)
	require.NoError(t, err)
	resp.Body.Close()

	resp, err = client.Get(s.Server.URL + "/auth/me")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

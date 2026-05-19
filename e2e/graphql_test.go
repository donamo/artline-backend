package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type gqlRequest struct {
	Query     string `json:"query"`
	Variables any    `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   json.RawMessage  `json:"data"`
	Errors []gqlError       `json:"errors,omitempty"`
}

type gqlError struct {
	Message string `json:"message"`
}

func gqlDo(t *testing.T, client *http.Client, url, query string, variables any) gqlResponse {
	t.Helper()
	body, _ := json.Marshal(gqlRequest{Query: query, Variables: variables})
	resp, err := client.Post(url+"/graphql", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	var result gqlResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func TestGraphQL_Me_Unauthenticated(t *testing.T) {
	s := NewSuite(t)

	result := gqlDo(t, http.DefaultClient, s.Server.URL, `query { me { id email } }`, nil)

	assert.Nil(t, result.Errors)
	assert.JSONEq(t, `{"me": null}`, string(result.Data))
}

func TestGraphQL_Me_Authenticated(t *testing.T) {
	s := NewSuite(t)
	client, user := s.LoginAs(t, "google-sub-gql", "gql@test.com")

	result := gqlDo(t, client, s.Server.URL, `query { me { id email displayName } }`, nil)

	require.Empty(t, result.Errors)
	var data struct {
		Me struct {
			ID          string  `json:"id"`
			Email       string  `json:"email"`
			DisplayName *string `json:"displayName"`
		} `json:"me"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	assert.Equal(t, user.ID.String(), data.Me.ID)
	assert.Equal(t, "gql@test.com", data.Me.Email)
}

func TestGraphQL_MyCreativeProjects_Unauthenticated(t *testing.T) {
	s := NewSuite(t)

	result := gqlDo(t, http.DefaultClient, s.Server.URL, `query { myCreativeProjects { id title } }`, nil)

	require.Len(t, result.Errors, 1)
	assert.Equal(t, "unauthorized", result.Errors[0].Message)
}

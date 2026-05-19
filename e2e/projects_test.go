package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const createProjectMutation = `
mutation CreateProject($input: CreateCreativeProjectInput!) {
  createCreativeProject(input: $input) {
    id title description startYear startMonth lyrics creationMethod
    links { id platform url label sortOrder }
  }
}`

const updateProjectMutation = `
mutation UpdateProject($id: ID!, $input: UpdateCreativeProjectInput!) {
  updateCreativeProject(id: $id, input: $input) {
    id title description startYear startMonth lyrics creationMethod
    links { id platform url label sortOrder }
  }
}`

const deleteProjectMutation = `
mutation DeleteProject($id: ID!) {
  deleteCreativeProject(id: $id)
}`

const myProjectsQuery = `
query {
  myCreativeProjects {
    id title startYear startMonth
    links { platform url }
  }
}`

const getProjectQuery = `
query GetProject($id: ID!) {
  creativeProject(id: $id) {
    id title description startYear startMonth lyrics creationMethod
    links { platform url label sortOrder }
  }
}`

func TestProject_Create(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-create", "create@test.com")

	result := gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{
			"title":      "Első dal",
			"startYear":  2026,
			"startMonth": 5,
			"lyrics":     "Valami dalszöveg",
			"links": []map[string]any{
				{"platform": "YOUTUBE", "url": "https://youtube.com/watch?v=abc"},
				{"platform": "SPOTIFY", "url": "https://open.spotify.com/track/xyz"},
			},
		},
	})

	require.Empty(t, result.Errors)
	var data struct {
		CreateCreativeProject struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			StartYear  int    `json:"startYear"`
			StartMonth int    `json:"startMonth"`
			Lyrics     string `json:"lyrics"`
			Links      []struct {
				Platform string `json:"platform"`
				URL      string `json:"url"`
			} `json:"links"`
		} `json:"createCreativeProject"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	p := data.CreateCreativeProject
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, "Első dal", p.Title)
	assert.Equal(t, 2026, p.StartYear)
	assert.Equal(t, 5, p.StartMonth)
	assert.Equal(t, "Valami dalszöveg", p.Lyrics)
	require.Len(t, p.Links, 2)
	assert.Equal(t, "YOUTUBE", p.Links[0].Platform)
	assert.Equal(t, "SPOTIFY", p.Links[1].Platform)
}

func TestProject_Create_Unauthenticated(t *testing.T) {
	s := NewSuite(t)

	result := gqlDo(t, http.DefaultClient, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "X", "startYear": 2026, "startMonth": 1},
	})

	require.Len(t, result.Errors, 1)
	assert.Equal(t, "unauthorized", result.Errors[0].Message)
}

func TestProject_List(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-list", "list@test.com")

	gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "Dal 1", "startYear": 2025, "startMonth": 3},
	})
	gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "Dal 2", "startYear": 2026, "startMonth": 1},
	})

	result := gqlDo(t, client, s.Server.URL, myProjectsQuery, nil)

	require.Empty(t, result.Errors)
	var data struct {
		MyCreativeProjects []struct {
			Title string `json:"title"`
		} `json:"myCreativeProjects"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	// újabb előbb (2026 > 2025)
	require.Len(t, data.MyCreativeProjects, 2)
	assert.Equal(t, "Dal 2", data.MyCreativeProjects[0].Title)
	assert.Equal(t, "Dal 1", data.MyCreativeProjects[1].Title)
}

func TestProject_List_OnlyOwn(t *testing.T) {
	s := NewSuite(t)
	client1, _ := s.LoginAs(t, "sub-own1", "own1@test.com")
	client2, _ := s.LoginAs(t, "sub-own2", "own2@test.com")

	gqlDo(t, client1, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "User1 dal", "startYear": 2026, "startMonth": 1},
	})

	result := gqlDo(t, client2, s.Server.URL, myProjectsQuery, nil)

	require.Empty(t, result.Errors)
	var data struct {
		MyCreativeProjects []any `json:"myCreativeProjects"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	assert.Empty(t, data.MyCreativeProjects)
}

func TestProject_Update(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-update", "update@test.com")

	created := gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{
			"title": "Eredeti cím", "startYear": 2025, "startMonth": 6,
			"links": []map[string]any{{"platform": "YOUTUBE", "url": "https://youtube.com/old"}},
		},
	})
	require.Empty(t, created.Errors)
	var createdData struct {
		CreateCreativeProject struct{ ID string `json:"id"` } `json:"createCreativeProject"`
	}
	require.NoError(t, json.Unmarshal(created.Data, &createdData))
	id := createdData.CreateCreativeProject.ID

	result := gqlDo(t, client, s.Server.URL, updateProjectMutation, map[string]any{
		"id": id,
		"input": map[string]any{
			"title": "Módosított cím",
			"links": []map[string]any{
				{"platform": "SPOTIFY", "url": "https://open.spotify.com/new"},
			},
		},
	})

	require.Empty(t, result.Errors)
	var data struct {
		UpdateCreativeProject struct {
			Title string `json:"title"`
			Links []struct {
				Platform string `json:"platform"`
			} `json:"links"`
		} `json:"updateCreativeProject"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	assert.Equal(t, "Módosított cím", data.UpdateCreativeProject.Title)
	require.Len(t, data.UpdateCreativeProject.Links, 1)
	assert.Equal(t, "SPOTIFY", data.UpdateCreativeProject.Links[0].Platform)
}

func TestProject_Update_OtherUser(t *testing.T) {
	s := NewSuite(t)
	client1, _ := s.LoginAs(t, "sub-upd-owner", "upd-owner@test.com")
	client2, _ := s.LoginAs(t, "sub-upd-other", "upd-other@test.com")

	created := gqlDo(t, client1, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "Owner dal", "startYear": 2026, "startMonth": 1},
	})
	var createdData struct {
		CreateCreativeProject struct{ ID string `json:"id"` } `json:"createCreativeProject"`
	}
	require.NoError(t, json.Unmarshal(created.Data, &createdData))

	result := gqlDo(t, client2, s.Server.URL, updateProjectMutation, map[string]any{
		"id":    createdData.CreateCreativeProject.ID,
		"input": map[string]any{"title": "Hacker cím"},
	})

	require.Len(t, result.Errors, 1)
	assert.Equal(t, "not found", result.Errors[0].Message)
}

func TestProject_Delete(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-delete", "delete@test.com")

	created := gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "Törlendő", "startYear": 2024, "startMonth": 12},
	})
	var createdData struct {
		CreateCreativeProject struct{ ID string `json:"id"` } `json:"createCreativeProject"`
	}
	require.NoError(t, json.Unmarshal(created.Data, &createdData))
	id := createdData.CreateCreativeProject.ID

	del := gqlDo(t, client, s.Server.URL, deleteProjectMutation, map[string]any{"id": id})
	require.Empty(t, del.Errors)

	get := gqlDo(t, client, s.Server.URL, getProjectQuery, map[string]any{"id": id})
	require.Len(t, get.Errors, 1)
	assert.Equal(t, "not found", get.Errors[0].Message)
}

func TestProject_Delete_OtherUser(t *testing.T) {
	s := NewSuite(t)
	client1, _ := s.LoginAs(t, "sub-del-owner", "del-owner@test.com")
	client2, _ := s.LoginAs(t, "sub-del-other", "del-other@test.com")

	created := gqlDo(t, client1, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "Protected", "startYear": 2026, "startMonth": 2},
	})
	var createdData struct {
		CreateCreativeProject struct{ ID string `json:"id"` } `json:"createCreativeProject"`
	}
	require.NoError(t, json.Unmarshal(created.Data, &createdData))

	result := gqlDo(t, client2, s.Server.URL, deleteProjectMutation, map[string]any{
		"id": createdData.CreateCreativeProject.ID,
	})

	require.Len(t, result.Errors, 1)
	assert.Equal(t, "not found", result.Errors[0].Message)
}

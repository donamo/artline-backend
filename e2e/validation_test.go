package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_EmptyTitle(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-val-1", "val1@test.com")

	result := gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "", "startYear": 2026, "startMonth": 1},
	})

	require.Len(t, result.Errors, 1)
	assert.Equal(t, "title is required", result.Errors[0].Message)
}

func TestValidation_InvalidYear(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-val-2", "val2@test.com")

	result := gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "Dal", "startYear": 1900, "startMonth": 1},
	})

	require.Len(t, result.Errors, 1)
	assert.Equal(t, "startYear must be between 1950 and 2100", result.Errors[0].Message)
}

func TestValidation_InvalidMonth(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-val-3", "val3@test.com")

	result := gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "Dal", "startYear": 2026, "startMonth": 13},
	})

	require.Len(t, result.Errors, 1)
	assert.Equal(t, "startMonth must be between 1 and 12", result.Errors[0].Message)
}

func TestValidation_InvalidURL(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-val-4", "val4@test.com")

	result := gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{
			"title": "Dal", "startYear": 2026, "startMonth": 1,
			"links": []map[string]any{
				{"platform": "YOUTUBE", "url": "nem-url"},
			},
		},
	})

	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "invalid URL")
}

func TestValidation_Update_InvalidURL(t *testing.T) {
	s := NewSuite(t)
	client, _ := s.LoginAs(t, "sub-val-5", "val5@test.com")

	created := gqlDo(t, client, s.Server.URL, createProjectMutation, map[string]any{
		"input": map[string]any{"title": "Dal", "startYear": 2026, "startMonth": 1},
	})
	require.Empty(t, created.Errors)
	var createdData struct {
		CreateCreativeProject struct{ ID string `json:"id"` } `json:"createCreativeProject"`
	}
	require.NoError(t, json.Unmarshal(created.Data, &createdData))

	result := gqlDo(t, client, s.Server.URL, updateProjectMutation, map[string]any{
		"id": createdData.CreateCreativeProject.ID,
		"input": map[string]any{
			"links": []map[string]any{
				{"platform": "SPOTIFY", "url": "ftp://invalid"},
			},
		},
	})

	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "invalid URL")
}

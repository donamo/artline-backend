package graph

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/donamo/artline-backend/internal/graph/model"
)

func ptr[T any](v T) *T { return &v }

func TestValidateCreateInput_OK(t *testing.T) {
	input := model.CreateCreativeProjectInput{
		Title:      "Első dal",
		StartYear:  2026,
		StartMonth: 5,
	}
	assert.NoError(t, validateCreateInput(input))
}

func TestValidateCreateInput_EmptyTitle(t *testing.T) {
	input := model.CreateCreativeProjectInput{Title: "  ", StartYear: 2026, StartMonth: 1}
	assert.EqualError(t, validateCreateInput(input), "title is required")
}

func TestValidateCreateInput_TitleTooLong(t *testing.T) {
	input := model.CreateCreativeProjectInput{
		Title:      strings.Repeat("a", 151),
		StartYear:  2026,
		StartMonth: 1,
	}
	assert.EqualError(t, validateCreateInput(input), "title must be at most 150 characters")
}

func TestValidateCreateInput_YearTooLow(t *testing.T) {
	input := model.CreateCreativeProjectInput{Title: "X", StartYear: 1949, StartMonth: 1}
	assert.EqualError(t, validateCreateInput(input), "startYear must be between 1950 and 2100")
}

func TestValidateCreateInput_YearTooHigh(t *testing.T) {
	input := model.CreateCreativeProjectInput{Title: "X", StartYear: 2101, StartMonth: 1}
	assert.EqualError(t, validateCreateInput(input), "startYear must be between 1950 and 2100")
}

func TestValidateCreateInput_MonthZero(t *testing.T) {
	input := model.CreateCreativeProjectInput{Title: "X", StartYear: 2026, StartMonth: 0}
	assert.EqualError(t, validateCreateInput(input), "startMonth must be between 1 and 12")
}

func TestValidateCreateInput_MonthThirteen(t *testing.T) {
	input := model.CreateCreativeProjectInput{Title: "X", StartYear: 2026, StartMonth: 13}
	assert.EqualError(t, validateCreateInput(input), "startMonth must be between 1 and 12")
}

func TestValidateCreateInput_DescriptionTooLong(t *testing.T) {
	input := model.CreateCreativeProjectInput{
		Title:       "X",
		StartYear:   2026,
		StartMonth:  1,
		Description: ptr(strings.Repeat("a", 10001)),
	}
	assert.EqualError(t, validateCreateInput(input), "description is too long")
}

func TestValidateCreateInput_LyricsTooLong(t *testing.T) {
	input := model.CreateCreativeProjectInput{
		Title:      "X",
		StartYear:  2026,
		StartMonth: 1,
		Lyrics:     ptr(strings.Repeat("a", 50001)),
	}
	assert.EqualError(t, validateCreateInput(input), "lyrics is too long")
}

func TestValidateCreateInput_CreationMethodTooLong(t *testing.T) {
	input := model.CreateCreativeProjectInput{
		Title:          "X",
		StartYear:      2026,
		StartMonth:     1,
		CreationMethod: ptr(strings.Repeat("a", 20001)),
	}
	assert.EqualError(t, validateCreateInput(input), "creationMethod is too long")
}

func TestValidateCreateInput_InvalidURL(t *testing.T) {
	input := model.CreateCreativeProjectInput{
		Title:      "X",
		StartYear:  2026,
		StartMonth: 1,
		Links: []*model.CreativeProjectLinkInput{
			{Platform: "YOUTUBE", URL: "not-a-url"},
		},
	}
	assert.ErrorContains(t, validateCreateInput(input), "invalid URL")
}

func TestValidateCreateInput_ValidURL(t *testing.T) {
	input := model.CreateCreativeProjectInput{
		Title:      "X",
		StartYear:  2026,
		StartMonth: 1,
		Links: []*model.CreativeProjectLinkInput{
			{Platform: "YOUTUBE", URL: "https://youtube.com/watch?v=abc"},
		},
	}
	assert.NoError(t, validateCreateInput(input))
}

func TestValidateUpdateInput_EmptyTitle(t *testing.T) {
	input := model.UpdateCreativeProjectInput{Title: ptr("")}
	assert.EqualError(t, validateUpdateInput(input), "title is required")
}

func TestValidateUpdateInput_NilTitle_OK(t *testing.T) {
	// nil title means "don't change" — should not validate
	input := model.UpdateCreativeProjectInput{}
	assert.NoError(t, validateUpdateInput(input))
}

func TestValidateUpdateInput_InvalidMonth(t *testing.T) {
	input := model.UpdateCreativeProjectInput{StartMonth: ptr(0)}
	assert.EqualError(t, validateUpdateInput(input), "startMonth must be between 1 and 12")
}

func TestValidateURL(t *testing.T) {
	cases := []struct {
		url   string
		valid bool
	}{
		{"https://youtube.com/watch?v=abc", true},
		{"http://soundcloud.com/track", true},
		{"not-a-url", false},
		{"ftp://example.com", false},
		{"https://", false},
		{"", false},
	}
	for _, c := range cases {
		err := validateURL(c.url)
		if c.valid {
			assert.NoError(t, err, c.url)
		} else {
			assert.Error(t, err, c.url)
		}
	}
}

package graph

import (
	"errors"
	"net/url"
	"strings"

	"github.com/donamo/artline-backend/internal/graph/model"
)

func validateCreateInput(input model.CreateCreativeProjectInput) error {
	if err := validateTitle(input.Title); err != nil {
		return err
	}
	if err := validateYear(input.StartYear); err != nil {
		return err
	}
	if err := validateMonth(input.StartMonth); err != nil {
		return err
	}
	if input.Description != nil {
		if err := validateMaxLen("description", *input.Description, 10000); err != nil {
			return err
		}
	}
	if input.Lyrics != nil {
		if err := validateMaxLen("lyrics", *input.Lyrics, 50000); err != nil {
			return err
		}
	}
	if input.CreationMethod != nil {
		if err := validateMaxLen("creationMethod", *input.CreationMethod, 20000); err != nil {
			return err
		}
	}
	for _, l := range input.Links {
		if err := validateURL(l.URL); err != nil {
			return err
		}
	}
	return nil
}

func validateUpdateInput(input model.UpdateCreativeProjectInput) error {
	if input.Title != nil {
		if err := validateTitle(*input.Title); err != nil {
			return err
		}
	}
	if input.StartYear != nil {
		if err := validateYear(*input.StartYear); err != nil {
			return err
		}
	}
	if input.StartMonth != nil {
		if err := validateMonth(*input.StartMonth); err != nil {
			return err
		}
	}
	if input.Description != nil {
		if err := validateMaxLen("description", *input.Description, 10000); err != nil {
			return err
		}
	}
	if input.Lyrics != nil {
		if err := validateMaxLen("lyrics", *input.Lyrics, 50000); err != nil {
			return err
		}
	}
	if input.CreationMethod != nil {
		if err := validateMaxLen("creationMethod", *input.CreationMethod, 20000); err != nil {
			return err
		}
	}
	for _, l := range input.Links {
		if err := validateURL(l.URL); err != nil {
			return err
		}
	}
	return nil
}

func validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
	}
	if len(title) > 150 {
		return errors.New("title must be at most 150 characters")
	}
	return nil
}

func validateYear(year int) error {
	if year < 1950 || year > 2100 {
		return errors.New("startYear must be between 1950 and 2100")
	}
	return nil
}

func validateMonth(month int) error {
	if month < 1 || month > 12 {
		return errors.New("startMonth must be between 1 and 12")
	}
	return nil
}

func validateMaxLen(field, value string, max int) error {
	if len(value) > max {
		return errors.New(field + " is too long")
	}
	return nil
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("invalid URL: " + raw)
	}
	return nil
}

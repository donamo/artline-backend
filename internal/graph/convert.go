package graph

import (
	"database/sql"

	dbsqlc "github.com/donamo/artline-backend/internal/db"
	"github.com/donamo/artline-backend/internal/graph/model"
)

func toModelProject(p dbsqlc.CreativeProject, links []dbsqlc.CreativeProjectLink) *model.CreativeProject {
	m := &model.CreativeProject{
		ID:         p.ID.String(),
		Title:      p.Title,
		StartYear:  int(p.StartYear),
		StartMonth: int(p.StartMonth),
		CreatedAt:  p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  p.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Links:      make([]*model.CreativeProjectLink, 0, len(links)),
	}
	if p.Description.Valid {
		m.Description = &p.Description.String
	}
	if p.Lyrics.Valid {
		m.Lyrics = &p.Lyrics.String
	}
	if p.CreationMethod.Valid {
		m.CreationMethod = &p.CreationMethod.String
	}
	for _, l := range links {
		m.Links = append(m.Links, toModelLink(l))
	}
	return m
}

func toModelLink(l dbsqlc.CreativeProjectLink) *model.CreativeProjectLink {
	ml := &model.CreativeProjectLink{
		ID:        l.ID.String(),
		Platform:  model.CreativeProjectLinkPlatform(l.Platform),
		URL:       l.Url,
		SortOrder: int(l.SortOrder),
	}
	if l.Label.Valid {
		ml.Label = &l.Label.String
	}
	return ml
}

func nullStr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

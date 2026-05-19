package graph

import (
	"database/sql"

	dbsqlc "github.com/donamo/artline-backend/internal/db"
)

type Resolver struct {
	db      *sql.DB
	queries *dbsqlc.Queries
}

func NewResolver(db *sql.DB) *Resolver {
	return &Resolver{
		db:      db,
		queries: dbsqlc.New(db),
	}
}

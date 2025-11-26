package pr

import "github.com/jmoiron/sqlx"

type PRRepository struct {
	db *sqlx.DB
}

func NewPRRepository(db *sqlx.DB) *PRRepository {
	return &PRRepository{db: db}
}

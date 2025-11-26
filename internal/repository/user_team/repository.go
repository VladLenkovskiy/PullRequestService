package user_team

import "github.com/jmoiron/sqlx"

type UserTeamRepository struct {
	db *sqlx.DB
}

func NewUserTeamRepository(db *sqlx.DB) *UserTeamRepository {
	return &UserTeamRepository{db: db}
}

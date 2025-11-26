package user_team

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pr-service/internal/domain"

	"github.com/jmoiron/sqlx"
)

func (r *UserTeamRepository) CreateTeam(ctx context.Context, team domain.Team) error {
	const queryCreateTeam = `
		INSERT INTO teams (name) 
		SELECT $1::text 
		WHERE NOT EXISTS (SELECT 1 FROM teams WHERE name = $1::text)`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, queryCreateTeam, team.Name)

	err = r.createUsers(ctx, tx, team)
	if err != nil {
		return fmt.Errorf("createUser: %w", err)

	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *UserTeamRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	const query = "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1::text"
	var dbUser userDB

	err := r.db.GetContext(ctx, &dbUser, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("query user: %w", err)
	}

	user := dbUser.toDomain()

	return &user, nil
}

func (r *UserTeamRepository) GetUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	const query = `SELECT user_id, username, team_name, is_active 
				   FROM users 
				   WHERE team_name = $1`

	var dbUser []userDB

	err := r.db.SelectContext(ctx, &dbUser, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("query users by team: %w", err)
	}

	result := make([]domain.User, 0, len(dbUser))
	for _, db := range dbUser {
		result = append(result, db.toDomain())
	}

	return result, nil
}

func (r *UserTeamRepository) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	const query = "SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1"

	var userDb []userDB

	err := r.db.SelectContext(ctx, &userDb, query, name)
	if err != nil {
		return nil, fmt.Errorf("query users by name: %w", err)
	}

	if len(userDb) == 0 {
		return nil, nil
	}
	users := make([]domain.User, 0, len(userDb))
	for _, db := range userDb {
		users = append(users, db.toDomain())
	}

	return &domain.Team{
		Name:    name,
		Members: users,
	}, nil
}

func (r *UserTeamRepository) SetUserActive(ctx context.Context, req domain.ActivateUserRequest) error {
	result, err := r.db.ExecContext(ctx, "UPDATE users SET is_active = $1 WHERE user_id = $2",
		req.IsActive,
		req.UserID)

	if err != nil {
		return fmt.Errorf("update user active status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *UserTeamRepository) createUsers(ctx context.Context, tx *sqlx.Tx, team domain.Team) error {
	const query = `
INSERT INTO users (user_id, username, team_name, is_active)
VALUES (:user_id, :username, :team_name, :is_active)
ON CONFLICT (user_id) DO NOTHING`

	dbUsers := make([]userDB, 0, len(team.Members))

	for _, u := range team.Members {
		dbUsers = append(dbUsers, fromDomain(u, team.Name))
	}

	_, err := tx.NamedExecContext(ctx, query, dbUsers)
	if err != nil {
		return fmt.Errorf("insert users: %w", err)
	}

	return nil
}

func (r *UserTeamRepository) DeactivateByTeam(ctx context.Context, teamName string) error {
	const query = `UPDATE users 
				   SET is_active = FALSE 
				   WHERE team_name = $1;`

	_, err := r.db.ExecContext(ctx, query, teamName)
	if err != nil {
		return fmt.Errorf("deactivated: %w", err)
	}

	return nil
}

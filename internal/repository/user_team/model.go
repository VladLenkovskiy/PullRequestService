package user_team

import "pr-service/internal/domain"

type teamDB struct {
	Name string `db:"name"`
}
type userDB struct {
	ID       string `db:"user_id"`
	Username string `db:"username"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
}

func (u *userDB) toDomain() domain.User {
	return domain.User{
		ID:       u.ID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func fromDomain(user domain.User, teamName string) userDB {
	return userDB{
		ID:       user.ID,
		Username: user.Username,
		TeamName: teamName,
		IsActive: user.IsActive,
	}
}

func (t *teamDB) ToDomain(members []domain.User) *domain.Team {
	return &domain.Team{
		Name:    t.Name,
		Members: members,
	}
}

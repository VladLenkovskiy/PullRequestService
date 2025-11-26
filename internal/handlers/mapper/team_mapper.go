package mapper

import (
	"pr-service/internal/domain"
	"pr-service/internal/handlers/dto"
)

func TeamToResponse(team domain.Team) dto.CreateTeamOut {
	return dto.CreateTeamOut{
		Name:    team.Name,
		Members: UsersToDTO(team.Members, team.Name),
	}
}

func TeamFromRequest(req dto.CreateTeamIn) domain.Team {
	return domain.Team{
		Name:    req.Name,
		Members: UsersFromDTO(req.Members),
	}
}

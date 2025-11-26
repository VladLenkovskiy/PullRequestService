package service

import (
	"context"
	"fmt"
	"pr-service/internal/domain"
)

type TeamService struct {
	teamRepo UserTeamRepository
}

func NewTeamService(tr UserTeamRepository) *TeamService {
	return &TeamService{
		teamRepo: tr,
	}
}

func (s *TeamService) Create(ctx context.Context, team domain.Team) error {
	return s.teamRepo.CreateTeam(ctx, team)
}

func (s *TeamService) Get(ctx context.Context, name string) (domain.Team, error) {
	team, err := s.teamRepo.GetByName(ctx, name)

	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to get team: %w", err)
	}

	if team == nil {
		return domain.Team{}, domain.ErrNotFound
	}

	return *team, nil
}

func (s *TeamService) DeactivateTeam(ctx context.Context, teamName string) error {
	return s.teamRepo.DeactivateByTeam(ctx, teamName)
}

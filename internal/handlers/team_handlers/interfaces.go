package teamhand

import (
	"context"
	"pr-service/internal/domain"
)

type TeamService interface {
	Create(ctx context.Context, team domain.Team) error
	Get(ctx context.Context, name string) (domain.Team, error)
	DeactivateTeam(ctx context.Context, teamName string) error
}

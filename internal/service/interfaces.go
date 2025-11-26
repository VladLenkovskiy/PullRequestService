package service

import (
	"context"
	"pr-service/internal/domain"
)

type UserTeamRepository interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	GetUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	GetByName(ctx context.Context, name string) (*domain.Team, error)
	SetUserActive(ctx context.Context, req domain.ActivateUserRequest) error
	DeactivateByTeam(ctx context.Context, teamName string) error
}

type PRRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) error
	GetByID(ctx context.Context, id string) (*domain.PullRequest, error)
	UpdatePR(ctx context.Context, request domain.PullRequest) error
	ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	GetByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
	GetAllPRs(ctx context.Context) ([]domain.PullRequest, error)
}

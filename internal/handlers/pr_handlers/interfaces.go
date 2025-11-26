package prhand

import (
	"context"
	"pr-service/internal/domain"
)

type PRService interface {
	Create(ctx context.Context, request domain.PullRequestCreate) (domain.PullRequest, error)
	Get(ctx context.Context, id string) (*domain.PullRequest, error)
	Merge(ctx context.Context, id string) (*domain.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, error)
	GetByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
	GetAllPRs(ctx context.Context) ([]domain.PullRequest, error)
}

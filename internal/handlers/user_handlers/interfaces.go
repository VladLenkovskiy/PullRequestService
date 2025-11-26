package userhand

import (
	"context"
	"pr-service/internal/domain"
)

type UserService interface {
	SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
}

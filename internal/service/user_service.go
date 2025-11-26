package service

import (
	"context"
	"fmt"
	"pr-service/internal/domain"
)

type UserService struct {
	userRepo UserTeamRepository
	prRepo   PRRepository
}

func NewUserService(ur UserTeamRepository, pr PRRepository) *UserService {
	return &UserService{
		userRepo: ur,
		prRepo:   pr,
	}
}

func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}

	if user.IsActive == isActive {
		return user, nil
	}

	user.IsActive = isActive

	request := domain.ActivateUserRequest{
		UserID:   userID,
		IsActive: isActive,
	}

	if err := s.userRepo.SetUserActive(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to update user active status: %w", err)
	}

	return user, nil
}

func (s *UserService) GetByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	_, err := s.getByID(ctx, reviewerID)
	if err != nil {
		return nil, err
	}

	prs, err := s.prRepo.GetByReviewer(ctx, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}

	return prs, nil
}

func (s *UserService) getByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

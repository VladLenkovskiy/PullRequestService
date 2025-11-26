package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"pr-service/internal/domain"
)

type PRService struct {
	prRepo   PRRepository
	userRepo UserTeamRepository
}

func NewPRService(pr PRRepository, ur UserTeamRepository) *PRService {
	return &PRService{
		prRepo:   pr,
		userRepo: ur,
	}
}

func (s *PRService) Create(ctx context.Context, request domain.PullRequestCreate) (domain.PullRequest, error) {
	existsPr, err := s.prRepo.GetByID(ctx, request.ID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("failed to get exists pr")
	}
	if existsPr != nil {
		return domain.PullRequest{}, domain.ErrPRAlreadyExists
	}

	author, err := s.userRepo.GetUserByID(ctx, request.AuthorID)
	if err != nil || author == nil {
		return domain.PullRequest{}, domain.ErrNotFound
	}

	team, err := s.userRepo.GetByName(ctx, author.TeamName)
	if err != nil || team == nil {
		return domain.PullRequest{}, fmt.Errorf("author's team not found")
	}

	reviewers := s.selectReviewers(team.Members, request.AuthorID, 2)

	pr := domain.PullRequest{
		ID:                request.ID,
		Name:              request.Name,
		AuthorID:          request.AuthorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now(),
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return domain.PullRequest{}, fmt.Errorf("failed to create PR: %w", err)
	}

	return pr, nil
}

func (s *PRService) Get(ctx context.Context, id string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if pr == nil {
		return nil, domain.ErrNotFound
	}

	return pr, nil
}

func (s *PRService) Merge(ctx context.Context, id string) (*domain.PullRequest, error) {
	pr, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if pr.Status == domain.PRStatusMerged {
		return pr, nil
	}

	if err = pr.Merge(); err != nil {
		return nil, err
	}

	if err = s.prRepo.UpdatePR(ctx, *pr); err != nil {
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	return pr, nil
}

func (s *PRService) Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, error) {
	pr, err := s.Get(ctx, prID)
	if err != nil {
		return nil, err
	}

	if !pr.IsPROpen() {
		return nil, domain.ErrPRMerged
	}

	assigned := false
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			assigned = true
			break
		}
	}
	if !assigned {
		return nil, domain.ErrNotAssigned
	}

	author, err := s.userRepo.GetUserByID(ctx, pr.AuthorID)
	if err != nil || author == nil {
		return nil, fmt.Errorf("author not found")
	}

	team, err := s.userRepo.GetByName(ctx, author.TeamName)
	if err != nil || team == nil {
		return nil, fmt.Errorf("team not found")
	}

	newReviewerID, err := s.selectNewReviewer(team.Members, pr.AuthorID, pr.AssignedReviewers)
	if err != nil {
		return nil, err
	}

	if err := s.prRepo.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		return nil, fmt.Errorf("failed to reassign reviewer: %w", err)
	}

	for i, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			pr.AssignedReviewers[i] = newReviewerID
			break
		}
	}

	return pr, nil
}

func (s *PRService) selectNewReviewer(members []domain.User, authorID string, currentReviewers []string) (string, error) {
	reviewerMap := make(map[string]bool)
	for _, rid := range currentReviewers {
		reviewerMap[rid] = true
	}

	var candidates []domain.User
	for _, member := range members {
		if member.IsActive && member.ID != authorID && !reviewerMap[member.ID] {
			candidates = append(candidates, member)
		}
	}

	if len(candidates) == 0 {
		return "", domain.ErrNoCandidate
	}

	idx := rand.Intn(len(candidates))
	return candidates[idx].ID, nil
}

func (s *PRService) selectReviewers(members []domain.User, authorID string, maxCount int) []string {
	var candidates []domain.User
	for _, member := range members {
		if member.IsActive && member.ID != authorID {
			candidates = append(candidates, member)
		}
	}

	if len(candidates) == 0 {
		return []string{}
	}

	count := maxCount
	if len(candidates) < maxCount {
		count = len(candidates)
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = candidates[i].ID
	}

	return reviewers
}

func (s *PRService) GetByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	user, err := s.userRepo.GetUserByID(ctx, reviewerID)
	if err != nil || user == nil {
		return nil, domain.ErrNotFound
	}

	prs, err := s.prRepo.GetByReviewer(ctx, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}

	return prs, nil
}

func (s *PRService) GetAllPRs(ctx context.Context) ([]domain.PullRequest, error) {
	prs, err := s.prRepo.GetAllPRs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all PRs: %w", err)
	}
	return prs, nil
}

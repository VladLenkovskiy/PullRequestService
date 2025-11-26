package mapper

import (
	"pr-service/internal/domain"
	"pr-service/internal/handlers/dto"
)

func PRToResponse(pr domain.PullRequest) dto.CreatePullRequestOut {
	return dto.CreatePullRequestOut{
		ID:        pr.ID,
		Name:      pr.Name,
		AuthorID:  pr.AuthorID,
		Status:    dto.PRStatus(pr.Status),
		Reviewers: pr.AssignedReviewers,
	}
}

func PRsToResponse(pr []domain.PullRequest) []dto.CreatePullRequestOut {
	prResponses := make([]dto.CreatePullRequestOut, len(pr))
	for i, pr := range pr {
		prResponses[i] = PRToResponse(pr)
	}
	return prResponses
}

func PRFromRequest(req dto.CreatePullRequestIn) domain.PullRequest {
	return domain.PullRequest{
		ID:       req.ID,
		Name:     req.Name,
		AuthorID: req.AuthorID,
		Status:   domain.PRStatusOpen,
	}
}

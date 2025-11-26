package domain

import (
	"errors"
	"time"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequestCreate struct {
	ID       string
	Name     string
	AuthorID string
}

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

func (pr *PullRequest) Merge() error {
	if pr.Status == PRStatusMerged {
		return errors.New("pull request already merged")
	}
	pr.Status = PRStatusMerged

	now := time.Now()

	pr.MergedAt = &now
	return nil
}

func (pr *PullRequest) IsPROpen() bool {
	return pr.Status == PRStatusOpen
}

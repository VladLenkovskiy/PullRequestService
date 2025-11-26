package dto

import "pr-service/internal/domain"

type PRStatus = domain.PRStatus

type CreatePullRequestIn struct {
	ID       string `json:"pull_request_id" validate:"required"`
	Name     string `json:"pull_request_name" validate:"required"`
	AuthorID string `json:"author_id" validate:"required"`
}

type PullRequestWrapper struct {
	PR CreatePullRequestOut `json:"pr"`
}

type CreatePullRequestOut struct {
	ID        string   `json:"pull_request_id"`
	Name      string   `json:"pull_request_name"`
	AuthorID  string   `json:"author_id"`
	Status    PRStatus `json:"status"`
	Reviewers []string `json:"assigned_reviewers"`
}

type MergePullRequest struct {
	ID string `json:"pull_request_id" validate:"required"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldReviewerID string `json:"old_reviewer_id" validate:"required"`
}

type SetUserActiveIn struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type UserWrapper struct {
	User UserDTO `json:"user"`
}

type UserDTO struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name,omitempty"`
}

type GetUserReviewsOut struct {
	UserID       string                 `json:"user_id"`
	PullRequests []CreatePullRequestOut `json:"pull_requests"`
}

type CreateTeamIn struct {
	Name    string    `json:"team_name" validate:"required"`
	Members []UserDTO `json:"members" validate:"required,min=1,dive"`
}

type CreateTeamOut struct {
	Name    string    `json:"team_name"`
	Members []UserDTO `json:"members"`
}

type TeamWrapper struct {
	Team CreateTeamOut `json:"team"`
}

type DeactivateOut struct {
	Status string `json:"deactivated"`
}

package domain

import "errors"

const (
	ErrCodeTeamExists  = "TEAM_EXISTS"
	ErrCodePRExists    = "PR_EXISTS"
	ErrCodePRMerged    = "PR_MERGED"
	ErrCodeNotAssigned = "NOT_ASSIGNED"
	ErrCodeNoCandidate = "NO_CANDIDATE"
	ErrCodeNotFound    = "NOT_FOUND"
	ErrCodeInvalidData = "INVALID_DATA"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

var (
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrNotFound          = errors.New("resource not found")
	ErrPRAlreadyExists   = errors.New("PR id already exists")
	ErrPRMerged          = errors.New("cannot reassign on merged PR")
	ErrNotAssigned       = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate       = errors.New("no active replacement candidate in team")
)

func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}

package handlers

import (
	"fmt"

	"github.com/go-playground/validator"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

var validationMessages = map[string]string{
	"MergePullRequest.ID:required": "pull_request_id is required",

	"ReassignReviewerRequest.PullRequestID:required": "pull_request_id is required",
	"ReassignReviewerRequest.OldReviewerID:required": "old_reviewer_id is required",

	"CreatePullRequest.ID:required":       "id is required",
	"CreatePullRequest.Name:required":     "name is required",
	"CreatePullRequest.AuthorID:required": "author_id is required",

	"CreateTeamRequest.Name:required":    "team_name is required",
	"CreateTeamRequest.Members:required": "members are required",
	"CreateTeamRequest.Members:min":      "members are required",

	"SetActiveRequest.UserID:required": "user_id is required",
}

func GetValidationErrorMessage(err error) string {
	validationErrors := err.(validator.ValidationErrors)

	for _, e := range validationErrors {
		key := fmt.Sprintf("%s:%s", e.Field(), e.Tag())
		if msg, ok := validationMessages[key]; ok {
			return msg
		}
	}

	return "validation error"
}

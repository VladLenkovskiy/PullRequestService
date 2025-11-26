package pr

import (
	"pr-service/internal/domain"
	"time"
)

type prDB struct {
	ID        string          `db:"pr_id"`
	Name      string          `db:"pr_name"`
	AuthorID  string          `db:"author_id"`
	Status    domain.PRStatus `db:"status"`
	CreatedAt time.Time       `db:"created_at"`
	MergedAt  *time.Time      `db:"merged_at"`
}

type reviewerDB struct {
	PRID       string `db:"pr_id"`
	ReviewerID string `db:"reviewer_id"`
}

func (p prDB) toDomain(reviewers []string) domain.PullRequest {
	pr := domain.PullRequest{
		ID:                p.ID,
		Name:              p.Name,
		AuthorID:          p.AuthorID,
		Status:            p.Status,
		AssignedReviewers: reviewers,
		CreatedAt:         p.CreatedAt,
		MergedAt:          p.MergedAt,
	}

	return pr
}

func fromDomain(pr domain.PullRequest) prDB {
	dbPR := prDB{
		ID:        pr.ID,
		Name:      pr.Name,
		AuthorID:  pr.AuthorID,
		Status:    pr.Status,
		CreatedAt: pr.CreatedAt,
		MergedAt:  pr.MergedAt,
	}

	return dbPR
}

func toReviewerDB(prID string, reviewerIDs []string) []reviewerDB {
	reviewers := make([]reviewerDB, 0, len(reviewerIDs))

	for _, reviewerID := range reviewerIDs {
		reviewers = append(reviewers, reviewerDB{
			PRID:       prID,
			ReviewerID: reviewerID,
		})
	}

	return reviewers
}

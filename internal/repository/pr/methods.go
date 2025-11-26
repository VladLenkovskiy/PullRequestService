package pr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pr-service/internal/domain"
)

func (r *PRRepository) Create(ctx context.Context, pr domain.PullRequest) error {
	const (
		queryCreatePR = `INSERT INTO pull_requests (pr_id, pr_name, author_id, status, created_at) 
                     VALUES (:pr_id, :pr_name, :author_id, :status, :created_at)`
		queryAssignReviewer = `INSERT INTO pull_request_reviewers (pr_id, reviewer_id) VALUES (:pr_id, :reviewer_id)`
	)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	dbPR := fromDomain(pr)

	_, err = tx.NamedExecContext(ctx, queryCreatePR, dbPR)
	if err != nil {
		return fmt.Errorf("insert pr: %w", err)
	}

	if len(pr.AssignedReviewers) > 0 {
		reviewers := toReviewerDB(pr.ID, pr.AssignedReviewers)

		_, err = tx.NamedExecContext(ctx, queryAssignReviewer, reviewers)
		if err != nil {
			return fmt.Errorf("assign reviewers: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *PRRepository) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	const query = `SELECT pr_id, pr_name, author_id, status, created_at, merged_at 
                  FROM pull_requests WHERE pr_id = $1::text`

	var dbPR prDB

	err := r.db.GetContext(ctx, &dbPR, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("query pr: %w", err)
	}

	reviewers, err := r.getReviewers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get reviewers: %w", err)
	}

	result := dbPR.toDomain(reviewers)
	return &result, nil
}

func (r *PRRepository) UpdatePR(ctx context.Context, request domain.PullRequest) error {
	const queryUpdatePRStatus = `
		UPDATE pull_requests 
		SET status = $2, merged_at = $3
		WHERE pr_id = $1`

	_, err := r.db.ExecContext(ctx, queryUpdatePRStatus, request.ID, request.Status, request.MergedAt)
	if err != nil {
		return fmt.Errorf("update pr status: %w", err)
	}
	return nil
}

func (r *PRRepository) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	const queryRemoveReviewer = `DELETE FROM pull_request_reviewers WHERE pr_id = $1 AND reviewer_id = $2`
	const queryAssignReviewer = `INSERT INTO pull_request_reviewers (pr_id, reviewer_id) VALUES ($1::text, $2::text)`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, queryRemoveReviewer, prID, oldReviewerID)
	if err != nil {
		return fmt.Errorf("remove old reviewer: %w", err)
	}

	_, err = tx.ExecContext(ctx, queryAssignReviewer, prID, newReviewerID)
	if err != nil {
		return fmt.Errorf("assign new reviewer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *PRRepository) GetByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	const query = `
        SELECT pr.pr_id, pr.pr_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
        FROM pull_requests pr
        INNER JOIN pull_request_reviewers prr ON pr.pr_id = prr.pr_id
        WHERE prr.reviewer_id = $1
        ORDER BY pr.created_at DESC`

	var prsDB []prDB

	err := r.db.SelectContext(ctx, &prsDB, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("query PRs by reviewer: %w", err)
	}

	prs := make([]domain.PullRequest, len(prsDB))

	for i, dbPR := range prsDB {
		reviewers, err := r.getReviewers(ctx, dbPR.ID)
		if err != nil {
			return nil, fmt.Errorf("get reviewers for PR %s: %w", dbPR.ID, err)
		}
		prs[i] = dbPR.toDomain(reviewers)
	}

	return prs, nil
}

func (r *PRRepository) GetAllPRs(ctx context.Context) ([]domain.PullRequest, error) {
	const query = `SELECT pr_id, pr_name, author_id, status, created_at, merged_at 
				   FROM pull_requests 
				   ORDER BY created_at DESC`

	var prsDB []prDB

	err := r.db.SelectContext(ctx, &prsDB, query)
	if err != nil {
		return nil, fmt.Errorf("query all PRs: %w", err)
	}

	prs := make([]domain.PullRequest, len(prsDB))

	for i, dbPR := range prsDB {
		reviewers, err := r.getReviewers(ctx, dbPR.ID)
		if err != nil {
			return nil, fmt.Errorf("get reviewers for PR %s: %w", dbPR.ID, err)
		}

		prs[i] = dbPR.toDomain(reviewers)
	}

	return prs, nil
}

func (r *PRRepository) getReviewers(ctx context.Context, prID string) ([]string, error) {
	const query = `SELECT reviewer_id FROM pull_request_reviewers WHERE pr_id = $1`

	var reviewers []string

	err := r.db.SelectContext(ctx, &reviewers, query, prID)
	if err != nil {
		return nil, fmt.Errorf("query reviewers: %w", err)
	}

	return reviewers, nil
}

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pr_id VARCHAR(255) NOT NULL,
    reviewer_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (pr_id, reviewer_id),
    FOREIGN KEY (pr_id) REFERENCES pull_requests(pr_id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reviewers_pr ON pull_request_reviewers(pr_id);
CREATE INDEX IF NOT EXISTS idx_reviewers_user ON pull_request_reviewers(reviewer_id);
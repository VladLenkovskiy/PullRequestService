package integration

import (
	"context"
	"pr-service/internal/repository/pr"
	"pr-service/internal/repository/user_team"
	"testing"

	"github.com/jmoiron/sqlx"

	"pr-service/internal/domain"
	"pr-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbDSN, err := getPostgresDSN()
	require.NoError(t, err)

	ctx := context.Background()

	db, err := sqlx.Open("postgres", dbDSN)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, cleanupDatabase(db))

	userTeamRepo := user_team.NewUserTeamRepository(db)
	prRepo := pr.NewPRRepository(db)

	teamService := service.NewTeamService(userTeamRepo)
	prService := service.NewPRService(prRepo, userTeamRepo)

	members := []domain.User{
		{ID: "u20", Username: "dev1", IsActive: true},
		{ID: "u21", Username: "dev2", IsActive: true},
		{ID: "u22", Username: "dev3", IsActive: true},
	}

	err = teamService.Create(ctx, domain.Team{Name: "qa-team", Members: members})
	require.NoError(t, err)

	t.Run("create PR with reviewers", func(t *testing.T) {
		pr, err := prService.Create(ctx, domain.PullRequestCreate{ID: "pr-100", Name: "Add feature", AuthorID: "u20"})
		require.NoError(t, err)
		assert.Equal(t, "pr-100", pr.ID)
		assert.Equal(t, domain.PRStatusOpen, pr.Status)
		assert.LessOrEqual(t, len(pr.AssignedReviewers), 2)
		assert.Greater(t, len(pr.AssignedReviewers), 0)

		for _, reviewerID := range pr.AssignedReviewers {
			assert.NotEqual(t, "u20", reviewerID)
		}
	})

	t.Run("merge PR", func(t *testing.T) {
		_, err := prService.Create(ctx, domain.PullRequestCreate{ID: "pr-101", Name: "Fix bug", AuthorID: "u21"})
		require.NoError(t, err)

		mergedPR, err := prService.Merge(ctx, "pr-101")
		require.NoError(t, err)
		assert.Equal(t, domain.PRStatusMerged, mergedPR.Status)

		mergedPR2, err := prService.Merge(ctx, "pr-101")
		require.NoError(t, err)
		assert.Equal(t, domain.PRStatusMerged, mergedPR2.Status)
	})

	t.Run("reassign reviewer with enough candidates", func(t *testing.T) {
		require.NoError(t, cleanupDatabase(db))

		members = []domain.User{
			{ID: "u30", Username: "author", IsActive: true},
			{ID: "u31", Username: "reviewer1", IsActive: true},
			{ID: "u32", Username: "reviewer2", IsActive: true},
			{ID: "u33", Username: "reviewer3", IsActive: true},
		}

		err = teamService.Create(ctx, domain.Team{Name: "big-team", Members: members})
		require.NoError(t, err)

		createPR, err := prService.Create(ctx, domain.PullRequestCreate{ID: "pr-103", Name: "Refactor", AuthorID: "u30"})
		require.NoError(t, err)
		require.NotEmpty(t, createPR.AssignedReviewers)

		oldReviewerID := createPR.AssignedReviewers[0]

		reassignedPR, err := prService.Reassign(ctx, "pr-103", oldReviewerID)
		require.NoError(t, err)
		assert.NotContains(t, reassignedPR.AssignedReviewers, oldReviewerID)
	})

	t.Run("reassign reviewer - no candidates", func(t *testing.T) {
		require.NoError(t, cleanupDatabase(db))

		members = []domain.User{
			{ID: "u40", Username: "author", IsActive: true},
			{ID: "u41", Username: "reviewer1", IsActive: true},
			{ID: "u42", Username: "reviewer2", IsActive: true},
		}

		err = teamService.Create(ctx, domain.Team{Name: "small-team", Members: members})
		require.NoError(t, err)

		createPR, err := prService.Create(ctx, domain.PullRequestCreate{ID: "pr-104", Name: "Feature", AuthorID: "u40"})
		require.NoError(t, err)
		require.Len(t, createPR.AssignedReviewers, 2)

		oldReviewerID := createPR.AssignedReviewers[0]

		reassignedPR, err := prService.Reassign(ctx, "pr-104", oldReviewerID)

		assert.Error(t, err)
		assert.Nil(t, reassignedPR)
		assert.ErrorIs(t, err, domain.ErrNoCandidate)
	})

	t.Run("get PRs by reviewer", func(t *testing.T) {
		require.NoError(t, cleanupDatabase(db))

		members = []domain.User{
			{ID: "u50", Username: "reviewer", IsActive: true},
			{ID: "u51", Username: "author", IsActive: true},
		}
		err = teamService.Create(ctx, domain.Team{Name: "test-team", Members: members})
		require.NoError(t, err)

		createPR, err := prService.Create(ctx, domain.PullRequestCreate{ID: "pr-200", Name: "Test PR", AuthorID: "u51"})
		require.NoError(t, err)

		if len(createPR.AssignedReviewers) > 0 {
			reviewerID := createPR.AssignedReviewers[0]
			prs, err := prService.GetByReviewer(ctx, reviewerID)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(prs), 1)
		}
	})

	err = cleanupDatabase(db)
	require.NoError(t, err)
}

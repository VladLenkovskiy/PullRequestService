package integration

import (
	"context"
	"pr-service/internal/repository/pr"
	"pr-service/internal/repository/user_team"
	"testing"

	"github.com/jmoiron/sqlx"

	"pr-service/internal/domain"
	"pr-service/internal/service"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserIntegration(t *testing.T) {
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
	userService := service.NewUserService(userTeamRepo, prRepo)

	members := []domain.User{
		{ID: "u40", Username: "user1", IsActive: true},
	}

	err = teamService.Create(ctx, domain.Team{Name: "dev-team", Members: members})
	require.NoError(t, err)

	t.Run("set user active status", func(t *testing.T) {
		user, err := userService.SetActive(ctx, "u40", false)
		require.NoError(t, err)
		assert.False(t, user.IsActive)

		user, err = userService.SetActive(ctx, "u40", true)
		require.NoError(t, err)
		assert.True(t, user.IsActive)
	})

	t.Run("set active for non-existent user", func(t *testing.T) {
		user, err := userService.SetActive(ctx, "u999", false)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	err = cleanupDatabase(db)
	require.NoError(t, err)
}

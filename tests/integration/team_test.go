package integration

import (
	"context"
	"pr-service/internal/repository/user_team"
	"testing"

	"github.com/jmoiron/sqlx"

	"pr-service/internal/domain"
	"pr-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbDSN, err := getPostgresDSN()
	require.NoError(t, err)

	ctx := context.Background()

	db, err := sqlx.Open("postgres", dbDSN)
	require.NoError(t, err)
	defer db.Close()

	userTeamRepo := user_team.NewUserTeamRepository(db)
	teamService := service.NewTeamService(userTeamRepo)

	t.Run("create and get team", func(t *testing.T) {
		members := []domain.User{
			{ID: "u1", Username: "alice", IsActive: true},
			{ID: "u2", Username: "bob", IsActive: true},
		}

		err = teamService.Create(ctx, domain.Team{Name: "backend", Members: members})
		require.NoError(t, err)

		team, err := teamService.Get(ctx, "backend")
		assert.NotNil(t, team)
		require.NoError(t, err)

		assert.Equal(t, "backend", team.Name)
		assert.Len(t, team.Members, 2)

		fetchedTeam, err := teamService.Get(ctx, "backend")
		require.NoError(t, err)
		assert.Equal(t, "backend", fetchedTeam.Name)
		assert.Len(t, fetchedTeam.Members, 2)
	})

	t.Run("idempotent team creation", func(t *testing.T) {
		require.NoError(t, cleanupDatabase(db))

		members1 := []domain.User{
			{ID: "u10", Username: "alice", IsActive: true},
		}

		err = teamService.Create(ctx, domain.Team{Name: "frontend", Members: members1})
		require.NoError(t, err)

		team1, err := teamService.Get(ctx, "frontend")
		assert.NotNil(t, team1)
		require.NoError(t, err)

		assert.Len(t, team1.Members, 1)

		members2 := []domain.User{
			{ID: "u10", Username: "alice", IsActive: true},
			{ID: "u11", Username: "charlie", IsActive: true},
		}

		err = teamService.Create(ctx, domain.Team{Name: "frontend", Members: members2})
		require.NoError(t, err)

		team2, err := teamService.Get(ctx, "frontend")
		assert.NotNil(t, team2)
		require.NoError(t, err)

		assert.Len(t, team2.Members, 2)
	})

	err = cleanupDatabase(db)
	require.NoError(t, err)

}

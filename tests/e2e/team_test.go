package e2e

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"pr-service/internal/domain"
	"pr-service/internal/handlers/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTeam_E2E(t *testing.T) {
	teamReq := dto.CreateTeamIn{
		Name: "Backend Team",
		Members: []dto.UserDTO{
			{ID: "user1", Username: "Alice", IsActive: true},
			{ID: "user2", Username: "Bob", IsActive: true},
		},
	}

	body, err := json.Marshal(teamReq)
	require.NoError(t, err)

	resp, err := http.Post(
		host+"/team/add",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result dto.TeamWrapper
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "Backend Team", result.Team.Name)
	assert.Len(t, result.Team.Members, 2)
	assert.Equal(t, "Alice", result.Team.Members[0].Username)
	assert.Equal(t, "Bob", result.Team.Members[1].Username)
}

func TestCreateTeam_AddNewMembers_E2E(t *testing.T) {
	user := "user" + strconv.Itoa(rand.Int())
	teamReq := dto.CreateTeamIn{
		Name: "Backend Team",
		Members: []dto.UserDTO{
			{ID: user, Username: "Alice", IsActive: true},
		},
	}

	body, err := json.Marshal(teamReq)
	require.NoError(t, err)

	resp1, err := http.Post(host+"/team/add", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	user2 := "user" + strconv.Itoa(rand.Int())

	teamReq.Members = []dto.UserDTO{
		{ID: user2, Username: "Bob", IsActive: true},
	}
	body, err = json.Marshal(teamReq)
	require.NoError(t, err)

	resp2, err := http.Post(host+"/team/add", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusCreated, resp2.StatusCode)

	u, err := url.Parse(host + "/team/get")
	require.NoError(t, err)
	q := u.Query()
	q.Set("team_name", "Backend Team")
	u.RawQuery = q.Encode()

	resp3, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp3.Body.Close()

	assert.Equal(t, http.StatusOK, resp3.StatusCode)

	var result dto.CreateTeamOut
	err = json.NewDecoder(resp3.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "Backend Team", result.Name)

	expectedUserIDs := map[string]bool{
		user:  false,
		user2: false,
	}

	for _, member := range result.Members {
		if member.ID == user {
			expectedUserIDs[user] = true
			assert.Equal(t, "Alice", member.Username)
		}
		if member.ID == user2 {
			expectedUserIDs[user2] = true
			assert.Equal(t, "Bob", member.Username)
		}
	}

	assert.True(t, expectedUserIDs[user], "User1 should be present in the team")
	assert.True(t, expectedUserIDs[user2], "User2 should be present in the team")
}

func TestGetTeam_E2E(t *testing.T) {

	teamReq := dto.CreateTeamIn{
		Name: "QA Team",
		Members: []dto.UserDTO{
			{ID: "user3", Username: "Charlie", IsActive: true},
		},
	}

	body, err := json.Marshal(teamReq)
	require.NoError(t, err)

	resp, err := http.Post(host+"/team/add", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	u, err := url.Parse(host + "/team/get")
	require.NoError(t, err)
	q := u.Query()
	q.Set("team_name", "QA Team")
	u.RawQuery = q.Encode()

	resp2, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var result dto.CreateTeamOut
	err = json.NewDecoder(resp2.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "QA Team", result.Name)
	assert.Len(t, result.Members, 1)
	assert.Equal(t, "Charlie", result.Members[0].Username)
}

func TestGetTeam_NotFound_E2E(t *testing.T) {
	u, err := url.Parse(host + "/team/get")
	require.NoError(t, err)
	q := u.Query()
	q.Set("team_name", "NonExistent")
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodeNotFound, errResp.Error.Code)
}

func TestGetTeam_MissingParameter_E2E(t *testing.T) {
	resp, err := http.Get(host + "/team/get")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodeInvalidData, errResp.Error.Code)
}

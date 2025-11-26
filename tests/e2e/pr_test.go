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

func createTeamForPR(t *testing.T, serverURL string) (authorID, reviewerID string) {
	t.Helper()

	teamReq := dto.CreateTeamIn{
		Name: "payments",
		Members: []dto.UserDTO{
			{ID: "u1", Username: "Alice", IsActive: true},
			{ID: "u2", Username: "Bob", IsActive: true},
		},
	}
	body, err := json.Marshal(teamReq)
	require.NoError(t, err)

	resp, err := http.Post(serverURL+"/team/add", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	return "u1", "u2"
}

func TestCreatePullRequest_E2E(t *testing.T) {
	authorID, _ := createTeamForPR(t, host)

	prID := "pr-" + strconv.Itoa(rand.Int())
	req := dto.CreatePullRequestIn{
		ID:       prID,
		Name:     "Add new feature",
		AuthorID: authorID,
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)

	resp, err := http.Post(host+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var out dto.PullRequestWrapper
	err = json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	assert.Equal(t, prID, out.PR.ID)
	assert.Equal(t, "Add new feature", out.PR.Name)
	assert.Equal(t, authorID, out.PR.AuthorID)
}

func TestCreatePullRequest_AlreadyExists_E2E(t *testing.T) {
	authorID, _ := createTeamForPR(t, host)

	id := "pr-dup" + strconv.Itoa(rand.Int())
	req := dto.CreatePullRequestIn{
		ID:       id,
		Name:     "Initial",
		AuthorID: authorID,
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)

	resp1, err := http.Post(host+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	resp2, err := http.Post(host+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp2.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodePRExists, errResp.Error.Code)
}

func TestCreatePullRequest_InvalidBody_E2E(t *testing.T) {

	resp, err := http.Post(host+"/pullRequest/create", "application/json", bytes.NewBufferString(`{`))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodeInvalidData, errResp.Error.Code)
}

func TestMergePullRequest_E2E(t *testing.T) {

	authorID, _ := createTeamForPR(t, host)

	prID := "pr-" + strconv.Itoa(rand.Int())

	createReq := dto.CreatePullRequestIn{
		ID:       prID,
		Name:     "To be merged",
		AuthorID: authorID,
	}
	body, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(host+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	mergeReq := dto.MergePullRequest{ID: prID}
	body, err = json.Marshal(mergeReq)
	require.NoError(t, err)

	resp2, err := http.Post(host+"/pullRequest/merge", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var out dto.PullRequestWrapper
	err = json.NewDecoder(resp2.Body).Decode(&out)
	require.NoError(t, err)

	assert.Equal(t, prID, out.PR.ID)
	assert.Equal(t, domain.PRStatusMerged, out.PR.Status)
}

func TestMergePullRequest_NotFound_E2E(t *testing.T) {
	mergeReq := dto.MergePullRequest{ID: "unknown-pr"}
	body, err := json.Marshal(mergeReq)
	require.NoError(t, err)

	resp, err := http.Post(host+"/pullRequest/merge", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodeNotFound, errResp.Error.Code)
}

func TestReassignReviewer_E2E(t *testing.T) {
	_, reviewerID := createTeamForPR(t, host)

	prID := "pr-" + strconv.Itoa(rand.Int())

	createReq := dto.CreatePullRequestIn{
		ID:       prID,
		Name:     "Need new reviewer",
		AuthorID: "u1",
	}
	body, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(host+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	reassignReq := dto.ReassignReviewerRequest{
		PullRequestID: prID,
		OldReviewerID: reviewerID,
	}
	body, err = json.Marshal(reassignReq)
	require.NoError(t, err)

	resp2, err := http.Post(host+"/pullRequest/reassign", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp2.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.NotEmpty(t, errResp.Error.Code)
}

func TestReassignReviewer_NotFound_E2E(t *testing.T) {
	reassignReq := dto.ReassignReviewerRequest{
		PullRequestID: "unknown",
		OldReviewerID: "u1",
	}
	body, err := json.Marshal(reassignReq)
	require.NoError(t, err)

	resp, err := http.Post(host+"/pullRequest/reassign", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodeNotFound, errResp.Error.Code)
}

func TestGetStats_E2E(t *testing.T) {
	authorID, _ := createTeamForPR(t, host)

	for i := 1; i <= 2; i++ {
		prID := "pr-" + strconv.Itoa(rand.Int())

		req := dto.CreatePullRequestIn{
			ID:       prID,
			Name:     "PR stats ",
			AuthorID: authorID,
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)

		resp, err := http.Post(host+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	u, err := url.Parse(host + "/stats")
	require.NoError(t, err)

	resp, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var res struct {
		Total int                        `json:"total_pull_requests"`
		PRs   []dto.CreatePullRequestOut `json:"pull_requests"`
	}
	err = json.NewDecoder(resp.Body).Decode(&res)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, res.Total, 2)
	assert.Len(t, res.PRs, res.Total)
}

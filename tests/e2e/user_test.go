package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"pr-service/internal/domain"
	"pr-service/internal/handlers/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetUserActive_E2E(t *testing.T) {

	teamReq := dto.CreateTeamIn{
		Name: "payments",
		Members: []dto.UserDTO{
			{ID: "u1", Username: "Alice", IsActive: true},
		},
	}
	body, err := json.Marshal(teamReq)
	require.NoError(t, err)

	resp, err := http.Post(host+"/team/add", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	setReq := dto.SetUserActiveIn{
		UserID:   "u1",
		IsActive: false,
	}
	body, err = json.Marshal(setReq)
	require.NoError(t, err)

	resp2, err := http.Post(host+"/users/setIsActive", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var result dto.UserWrapper
	err = json.NewDecoder(resp2.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "u1", result.User.ID)
	assert.False(t, result.User.IsActive)
}

func TestSetUserActive_NotFound_E2E(t *testing.T) {

	setReq := dto.SetUserActiveIn{
		UserID:   "unknown",
		IsActive: false,
	}
	body, err := json.Marshal(setReq)
	require.NoError(t, err)

	resp, err := http.Post(host+"/users/setIsActive", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodeNotFound, errResp.Error.Code)
}

func TestSetUserActive_InvalidBody_E2E(t *testing.T) {

	resp, err := http.Post(host+"/users/setIsActive", "application/json", bytes.NewBufferString(`{`))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodeInvalidData, errResp.Error.Code)
}

func TestGetUserReviews_E2E(t *testing.T) {

	teamReq := dto.CreateTeamIn{
		Name: "payments",
		Members: []dto.UserDTO{
			{ID: "u1", Username: "Alice", IsActive: true},
			{ID: "u2", Username: "Bob", IsActive: true},
		},
	}
	body, err := json.Marshal(teamReq)
	require.NoError(t, err)

	resp, err := http.Post(host+"/team/add", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	u, err := url.Parse(host + "/users/getReview")
	require.NoError(t, err)
	q := u.Query()
	q.Set("user_id", "u2")
	u.RawQuery = q.Encode()

	resp2, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var out dto.GetUserReviewsOut
	err = json.NewDecoder(resp2.Body).Decode(&out)
	require.NoError(t, err)

	assert.Equal(t, "u2", out.UserID)
}

func TestGetUserReviews_MissingUserID_E2E(t *testing.T) {

	resp, err := http.Get(host + "/users/getReview")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp domain.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, domain.ErrCodeInvalidData, errResp.Error.Code)
}

func TestGetUserReviews_NotFound_E2E(t *testing.T) {

	u, err := url.Parse(host + "/users/getReview")
	require.NoError(t, err)
	q := u.Query()
	q.Set("user_id", "unknown")
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

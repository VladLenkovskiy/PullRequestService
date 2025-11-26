package prhand

import (
	"errors"

	"net/http"
	"pr-service/internal/domain"
	"pr-service/internal/handlers"
	"pr-service/internal/handlers/dto"
	"pr-service/internal/handlers/mapper"

	"github.com/go-chi/chi/v5"
)

type PRHandler struct {
	prService PRService
}

func NewPRHandler(prService PRService) *PRHandler {
	return &PRHandler{
		prService: prService,
	}
}

func (h *PRHandler) RegisterRoutes(r chi.Router) {
	r.Post("/pullRequest/create", h.CreatePullRequest)
	r.Post("/pullRequest/merge", h.MergePullRequest)
	r.Post("/pullRequest/reassign", h.ReassignReviewer)
	r.Get("/stats", h.GetStats)
}

func (h *PRHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	req, ok := handlers.DecodeAndValidate[dto.CreatePullRequestIn](w, r)
	if !ok {
		return
	}

	pullRequestCreate := domain.PullRequestCreate{
		ID:       req.ID,
		Name:     req.Name,
		AuthorID: req.AuthorID,
	}

	pr, err := h.prService.Create(r.Context(), pullRequestCreate)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlers.RespondError(w, http.StatusNotFound, domain.ErrCodeNotFound, domain.ErrNotFound.Error())
			return
		}

		if errors.Is(err, domain.ErrPRAlreadyExists) {
			handlers.RespondError(w, http.StatusConflict, domain.ErrCodePRExists, err.Error())
			return
		}

		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	response := dto.PullRequestWrapper{
		PR: mapper.PRToResponse(pr),
	}

	handlers.RespondJSON(w, http.StatusCreated, response)
}

func (h *PRHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	req, ok := handlers.DecodeAndValidate[dto.MergePullRequest](w, r)
	if !ok {
		return
	}

	pr, err := h.prService.Merge(r.Context(), req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlers.RespondError(w, http.StatusNotFound, domain.ErrCodeNotFound, domain.ErrNotFound.Error())
			return
		}

		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	response := dto.PullRequestWrapper{
		PR: mapper.PRToResponse(*pr),
	}

	handlers.RespondJSON(w, http.StatusOK, response)
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	req, ok := handlers.DecodeAndValidate[dto.ReassignReviewerRequest](w, r)
	if !ok {
		return
	}

	pr, err := h.prService.Reassign(r.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlers.RespondError(w, http.StatusNotFound, domain.ErrCodeNotFound, domain.ErrNotFound.Error())
			return
		}
		if errors.Is(err, domain.ErrPRMerged) {
			handlers.RespondError(w, http.StatusConflict, domain.ErrCodePRMerged, err.Error())
			return
		}
		if errors.Is(err, domain.ErrNotAssigned) {
			handlers.RespondError(w, http.StatusConflict, domain.ErrCodeNotAssigned, err.Error())
			return
		}
		if errors.Is(err, domain.ErrNoCandidate) {
			handlers.RespondError(w, http.StatusConflict, domain.ErrCodeNoCandidate, err.Error())
			return
		}
		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	response := dto.PullRequestWrapper{
		PR: mapper.PRToResponse(*pr),
	}

	handlers.RespondJSON(w, http.StatusOK, response)
}

func (h *PRHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	prs, err := h.prService.GetAllPRs(r.Context())
	if err != nil {
		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	prResponses := mapper.PRsToResponse(prs)

	response := map[string]interface{}{
		"total_pull_requests": len(prs),
		"pull_requests":       prResponses,
	}

	handlers.RespondJSON(w, http.StatusOK, response)
}

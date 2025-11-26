package userhand

import (
	"errors"
	"net/http"
	"pr-service/internal/domain"
	"pr-service/internal/handlers"
	"pr-service/internal/handlers/dto"
	"pr-service/internal/handlers/mapper"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/users/setIsActive", h.SetUserActive)
	r.Get("/users/getReview", h.GetUserReviews)
}

func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	req, ok := handlers.DecodeAndValidate[dto.SetUserActiveIn](w, r)
	if !ok {
		return
	}

	user, err := h.userService.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlers.RespondError(w, http.StatusNotFound, domain.ErrCodeNotFound, domain.ErrNotFound.Error())
			return
		}

		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	resp := dto.UserWrapper{
		User: mapper.UserToDTO(*user, user.TeamName),
	}
	handlers.RespondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		handlers.RespondError(w, http.StatusBadRequest, domain.ErrCodeInvalidData, "user_id is required")
		return
	}

	prs, err := h.userService.GetByReviewer(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlers.RespondError(w, http.StatusNotFound, domain.ErrCodeNotFound, domain.ErrNotFound.Error())
			return
		}

		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	items := mapper.PRsToResponse(prs)

	response := dto.GetUserReviewsOut{
		UserID:       userID,
		PullRequests: items,
	}

	handlers.RespondJSON(w, http.StatusOK, response)
}

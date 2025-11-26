package teamhand

import (
	"errors"
	"net/http"
	"pr-service/internal/domain"
	"pr-service/internal/handlers"
	"pr-service/internal/handlers/dto"
	"pr-service/internal/handlers/mapper"

	"github.com/go-chi/chi/v5"
)

type TeamHandler struct {
	teamService TeamService
}

func NewTeamHandler(teamService TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) RegisterRoutes(r chi.Router) {
	r.Post("/team/add", h.CreateTeam)
	r.Get("/team/get", h.GetTeam)
	r.Post("/deactivate", h.DeactivateTeam)
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	req, ok := handlers.DecodeAndValidate[dto.CreateTeamIn](w, r)
	if !ok {
		return
	}

	teamDomain := mapper.TeamFromRequest(req)

	err := h.teamService.Create(r.Context(), teamDomain)
	if err != nil {
		if errors.Is(err, domain.ErrTeamAlreadyExists) {
			handlers.RespondError(w, http.StatusConflict, domain.ErrCodeTeamExists, err.Error())
			return
		}

		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	resp := dto.TeamWrapper{
		Team: mapper.TeamToResponse(teamDomain),
	}

	handlers.RespondJSON(w, http.StatusCreated, resp)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("team_name")

	if err := handlers.Validate.Var(name, "required"); err != nil {
		handlers.RespondError(w, http.StatusBadRequest, domain.ErrCodeInvalidData, "team_name is required")
		return
	}

	team, err := h.teamService.Get(r.Context(), name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlers.RespondError(w, http.StatusNotFound, domain.ErrCodeNotFound, "team not found")
			return
		}
		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	handlers.RespondJSON(w, http.StatusOK, mapper.TeamToResponse(team))
}

func (h *TeamHandler) DeactivateTeam(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("team_name")

	if err := handlers.Validate.Var(name, "required"); err != nil {
		handlers.RespondError(w, http.StatusBadRequest, domain.ErrCodeInvalidData, "team_name is required")
		return
	}

	err := h.teamService.DeactivateTeam(r.Context(), name)
	if err != nil {
		handlers.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	handlers.RespondJSON(w, http.StatusOK, dto.DeactivateOut{Status: "deactivated"})
}

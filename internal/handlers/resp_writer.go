package handlers

import (
	"encoding/json"
	"net/http"
	"pr-service/internal/domain"
)

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, status int, code string, msg string) {
	errResp := domain.NewErrorResponse(code, msg)
	RespondJSON(w, status, errResp)
}

func DecodeAndValidate[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var data T
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		RespondError(w, http.StatusBadRequest, domain.ErrCodeInvalidData, "invalid request body")
		return data, false
	}

	if err := Validate.Struct(data); err != nil {
		msg := GetValidationErrorMessage(err)
		RespondError(w, http.StatusBadRequest, domain.ErrCodeInvalidData, msg)
		return data, false
	}

	return data, true
}

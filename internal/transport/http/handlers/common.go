package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	patterndomain "example.com/taskservice/internal/domain/pattern"
	taskdomain "example.com/taskservice/internal/domain/task"
	patternusecase "example.com/taskservice/internal/usecase/pattern"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, patterndomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, patternusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

var ErrValidation = errors.New("validation failed")

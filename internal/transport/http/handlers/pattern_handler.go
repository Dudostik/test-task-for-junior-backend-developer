package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	patterndomain "example.com/taskservice/internal/domain/pattern"
	patternusecase "example.com/taskservice/internal/usecase/pattern"
)

type PatternHandler struct {
	usecase patternusecase.Usecase
}

func NewPatternHandler(usecase patternusecase.Usecase) *PatternHandler {
	return &PatternHandler{usecase: usecase}
}

type createPatternRequest struct {
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	Type          patterndomain.PatternType  `json:"type"`
	IntervalDays  *int                       `json:"interval_days,omitempty"`
	DayOfMonth    *int                       `json:"day_of_month,omitempty"`
	SpecificDates []time.Time                `json:"specific_dates,omitempty"`
	EvenOddType   *patterndomain.EvenOddType `json:"even_odd_type,omitempty"`
	StartDate     *time.Time                 `json:"start_date,omitempty"`
	EndDate       *time.Time                 `json:"end_date,omitempty"`
	IsActive      bool                       `json:"is_active"`
}

func (h *PatternHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPatternRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), patternusecase.CreatePatternInput{
		Name:          req.Name,
		Description:   req.Description,
		Type:          req.Type,
		IntervalDays:  req.IntervalDays,
		DayOfMonth:    req.DayOfMonth,
		SpecificDates: req.SpecificDates,
		EvenOddType:   req.EvenOddType,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		IsActive:      req.IsActive,
	})

	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *PatternHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getPatternIdFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	pattern, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, pattern)
}

func (h *PatternHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getPatternIdFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req createPatternRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := patternusecase.UpdatePatternInput{
		Name:          &req.Name,
		Description:   &req.Description,
		Type:          &req.Type,
		IntervalDays:  req.IntervalDays,
		DayOfMonth:    req.DayOfMonth,
		SpecificDates: req.SpecificDates,
		EvenOddType:   req.EvenOddType,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		IsActive:      &req.IsActive,
	}

	updated, err := h.usecase.Update(r.Context(), id, input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *PatternHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getPatternIdFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PatternHandler) List(w http.ResponseWriter, r *http.Request) {
	patterns, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, patterns)
}

func getPatternIdFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing pattern id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid pattern id")
	}

	if id <= 0 {
		return 0, errors.New("invalid pattern id")
	}

	return id, nil
}

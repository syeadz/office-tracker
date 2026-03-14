package http

import (
	"encoding/json"
	"math"
	"net/http"

	"office/internal/api/dto"
	"office/internal/domain"
	"office/internal/service"
)

// EnvironmentHandler handles in-memory environmental data.
type EnvironmentHandler struct {
	environmentSvc *service.EnvironmentService
}

// NewEnvironmentHandler creates a new EnvironmentHandler.
func NewEnvironmentHandler(environmentSvc *service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{environmentSvc: environmentSvc}
}

// GetLatest returns the latest environmental reading and freshness metadata.
func (h *EnvironmentHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	response := dto.EnvironmentResponse{}
	reading, ok := h.environmentSvc.GetLatest()
	if ok {
		response = h.toResponse(reading)
	}

	writeJSON(w, http.StatusOK, response)
}

// UpdateLatest stores the latest environmental reading in memory.
func (h *EnvironmentHandler) UpdateLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.EnvironmentUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request")
		return
	}

	if math.IsNaN(req.TemperatureC) || math.IsInf(req.TemperatureC, 0) {
		writeErrorJSON(w, http.StatusBadRequest, "invalid temperature_c")
		return
	}

	reading := domain.EnvironmentReading{
		TemperatureC: req.TemperatureC,
		Timestamp:    req.Timestamp,
	}
	h.environmentSvc.Update(reading)

	latest, _ := h.environmentSvc.GetLatest()
	writeJSON(w, http.StatusOK, h.toResponse(latest))
}

func (h *EnvironmentHandler) toResponse(reading *domain.EnvironmentReading) dto.EnvironmentResponse {
	if reading == nil {
		return dto.EnvironmentResponse{}
	}

	fresh := h.environmentSvc.IsFresh(reading, 0)
	age := h.environmentSvc.Age(reading)
	timestamp := reading.Timestamp

	return dto.EnvironmentResponse{
		Available:    true,
		Fresh:        fresh,
		TemperatureC: reading.TemperatureC,
		Timestamp:    &timestamp,
		AgeSeconds:   int64(age.Seconds()),
	}
}

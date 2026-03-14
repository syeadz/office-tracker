package http

import (
	"encoding/json"
	"net/http"

	"office/internal/api/dto"
	"office/internal/domain"
	"office/internal/service"
)

// ESPHealthHandler manages in-memory ESP32 heartbeat data.
type ESPHealthHandler struct {
	healthSvc *service.ESPHealthService
}

// NewESPHealthHandler creates an ESPHealthHandler.
func NewESPHealthHandler(healthSvc *service.ESPHealthService) *ESPHealthHandler {
	return &ESPHealthHandler{healthSvc: healthSvc}
}

// Upsert stores latest health status for an ESP32 device.
func (h *ESPHealthHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.ESPHealthUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.UptimeSeconds < 0 {
		writeErrorJSON(w, http.StatusBadRequest, "uptime_seconds must be >= 0")
		return
	}
	if req.FreeHeapBytes < 0 {
		writeErrorJSON(w, http.StatusBadRequest, "free_heap_bytes must be >= 0")
		return
	}

	h.healthSvc.Update(domain.ESPHealthStatus{
		DeviceID:        req.DeviceID,
		UptimeSeconds:   req.UptimeSeconds,
		FreeHeapBytes:   req.FreeHeapBytes,
		WiFiConnected:   req.WiFiConnected,
		RSSI:            req.RSSI,
		IP:              req.IP,
		FirmwareVersion: req.FirmwareVersion,
		ResetReason:     req.ResetReason,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// List returns the latest health status for all ESP32 devices.
func (h *ESPHealthHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	items := h.healthSvc.GetAll()
	response := make([]dto.ESPHealthStatusResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.ESPHealthStatusResponse{
			DeviceID:        item.DeviceID,
			UptimeSeconds:   item.UptimeSeconds,
			FreeHeapBytes:   item.FreeHeapBytes,
			WiFiConnected:   item.WiFiConnected,
			RSSI:            item.RSSI,
			IP:              item.IP,
			FirmwareVersion: item.FirmwareVersion,
			ResetReason:     item.ResetReason,
			UpdatedAt:       item.UpdatedAt,
			Fresh:           h.healthSvc.IsFresh(item),
			AgeSeconds:      int64(h.healthSvc.Age(item).Seconds()),
		})
	}

	writeJSON(w, http.StatusOK, response)
}

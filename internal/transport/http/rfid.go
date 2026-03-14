package http

import (
	"encoding/json"
	"net/http"

	"office/internal/api/dto"
	"office/internal/service"
)

func RfidHandler(attSvc *service.AttendanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req dto.ScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("invalid request body", "err", err)
			writeErrorJSON(w, http.StatusBadRequest, "invalid request")
			return
		}

		result, err := attSvc.Scan(req.UID)
		if err != nil {
			log.Warn("scan failed", "uid", req.UID, "err", err)
			writeErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// ScanHistoryHandler returns the complete scan history (all scans, known and unknown)
func ScanHistoryHandler(attSvc *service.AttendanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			attSvc.ClearScanHistory()
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodGet {
			writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		scanHistory := attSvc.GetScanHistory()
		writeJSON(w, http.StatusOK, scanHistory)
	}
}

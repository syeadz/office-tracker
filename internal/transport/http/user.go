package http

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"office/internal/api/dto"
	"office/internal/query"
	"office/internal/service"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("invalid request body", "err", err)
		writeErrorJSON(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Name == "" || req.RFIDUID == "" {
		writeErrorJSON(w, http.StatusBadRequest, "name and rfid_uid are required")
		return
	}

	user, err := h.userSvc.CreateUser(req.Name, req.RFIDUID, req.DiscordID)
	if err != nil {
		log.Error("failed to create user", "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract ID from path
	idStr := r.URL.Path[len("/api/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.userSvc.GetUserByID(id)
	if err != nil {
		log.Error("failed to get user", "id", id, "err", err)
		writeErrorJSON(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract ID from path
	idStr := r.URL.Path[len("/api/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("invalid request body", "err", err)
		writeErrorJSON(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Name == "" {
		writeErrorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	user, err := h.userSvc.UpdateUser(id, req.Name, req.RFIDUID, req.DiscordID)
	if err != nil {
		log.Error("failed to update user", "id", id, "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract ID from path
	idStr := r.URL.Path[len("/api/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	err = h.userSvc.DeleteUser(id)
	if err != nil {
		log.Error("failed to delete user", "id", id, "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListUsers handles GET /api/users with query parameters
// Supports: ?search=&limit=&offset=&order=&sort_by=
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := query.UserFilter{}

	// Parse query parameters
	if search := r.URL.Query().Get("search"); search != "" {
		filter.NameLike = &search
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			writeErrorJSON(w, http.StatusBadRequest, "invalid limit")
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50 // default
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			writeErrorJSON(w, http.StatusBadRequest, "invalid offset")
			return
		}
		filter.Offset = offset
	}

	// Parse order parameter
	if order := r.URL.Query().Get("order"); order != "" {
		order = strings.ToLower(order)
		if order == "asc" || order == "desc" {
			filter.OrderBy = order
		} else {
			writeErrorJSON(w, http.StatusBadRequest, "invalid order - must be 'asc' or 'desc'")
			return
		}
	}

	// Parse sort_by parameter
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		sortBy = strings.ToLower(sortBy)
		if sortBy == "name" || sortBy == "created_at" {
			filter.SortBy = sortBy
		} else {
			writeErrorJSON(w, http.StatusBadRequest, "invalid sort_by - must be 'name' or 'created_at'")
			return
		}
	}

	users, err := h.userSvc.ListUsers(filter)
	if err != nil {
		log.Error("failed to list users", "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// DeleteUsers handles DELETE /api/users (bulk delete with filters)
func (h *UserHandler) DeleteUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := query.UserFilter{}

	// Parse query parameters
	if search := r.URL.Query().Get("search"); search != "" {
		filter.NameLike = &search
	}

	// Safety check - require at least one filter to prevent accidental delete all
	if filter.NameLike == nil {
		writeErrorJSON(w, http.StatusBadRequest, "filter required for bulk delete (e.g., ?search=name)")
		return
	}

	// Parse order parameter
	if order := r.URL.Query().Get("order"); order != "" {
		order = strings.ToLower(order)
		if order == "asc" || order == "desc" {
			filter.OrderBy = order
		} else {
			writeErrorJSON(w, http.StatusBadRequest, "invalid order - must be 'asc' or 'desc'")
			return
		}
	}

	count, err := h.userSvc.DeleteUsers(filter)
	if err != nil {
		log.Error("failed to delete users", "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to delete users")
		return
	}

	writeJSON(w, http.StatusOK, dto.DeleteResult{Deleted: count})
}

// ExportUsersCSV handles GET /api/users/export - exports users as CSV
func (h *UserHandler) ExportUsersCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filter := query.UserFilter{}

	if search := r.URL.Query().Get("search"); search != "" {
		filter.NameLike = &search
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50 // default
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		filter.Offset = offset
	}

	if order := r.URL.Query().Get("order"); order != "" {
		order = strings.ToLower(order)
		if order == "asc" || order == "desc" {
			filter.OrderBy = order
		} else {
			http.Error(w, "invalid order - must be 'asc' or 'desc'", http.StatusBadRequest)
			return
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		sortBy = strings.ToLower(sortBy)
		if sortBy == "name" || sortBy == "created_at" {
			filter.SortBy = sortBy
		} else {
			http.Error(w, "invalid sort_by - must be 'name' or 'created_at'", http.StatusBadRequest)
			return
		}
	}

	users, err := h.userSvc.ListUsersRaw(filter)
	if err != nil {
		log.Error("failed to list users", "err", err)
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=users.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Name", "RFID_UID", "DiscordID"})

	// Write data
	for _, u := range users {
		writer.Write([]string{u.Name, u.RFIDUID, u.DiscordID})
	}
}

// ImportUsersCSV handles POST /api/users/import - imports users from CSV
func (h *UserHandler) ImportUsersCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		http.Error(w, "failed to read CSV header", http.StatusBadRequest)
		return
	}

	// Validate header
	if len(header) < 2 || header[0] != "Name" || header[1] != "RFID_UID" {
		http.Error(w, "invalid CSV format. Expected: Name,RFID_UID,DiscordID", http.StatusBadRequest)
		return
	}

	imported := 0
	failed := 0

	// Read records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("error reading CSV record", "err", err)
			failed++
			continue
		}

		if len(record) < 2 {
			failed++
			continue
		}

		name := record[0]
		rfidUID := record[1]
		discordID := ""
		if len(record) > 2 {
			discordID = record[2]
		}

		_, err = h.userSvc.CreateUser(name, rfidUID, discordID)
		if err != nil {
			log.Error("failed to create user from CSV", "name", name, "err", err)
			failed++
			continue
		}

		imported++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"imported": imported,
		"failed":   failed,
	})
}

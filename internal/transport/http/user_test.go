package http_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"office/internal/api/dto"
	"office/internal/repository"
	"office/internal/service"
	httptransport "office/internal/transport/http"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  bool
	}{
		{
			name: "successful user creation",
			requestBody: dto.CreateUserRequest{
				Name:      "John Doe",
				RFIDUID:   "ABC123",
				DiscordID: "123456",
			},
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name: "missing name",
			requestBody: dto.CreateUserRequest{
				RFIDUID: "ABC123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name: "missing rfid_uid",
			requestBody: dto.CreateUserRequest{
				Name: "John Doe",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:           "invalid json",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.CreateUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse && w.Code == http.StatusCreated {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "John Doe", response["name"])
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed a user
	user := helpers.SeedUser(t, db, "Jane Doe", "XYZ789", "789456")

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		checkName      bool
	}{
		{
			name:           "get existing user",
			userID:         "1",
			expectedStatus: http.StatusOK,
			checkName:      true,
		},
		{
			name:           "user not found",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			checkName:      false,
		},
		{
			name:           "invalid user id",
			userID:         "abc",
			expectedStatus: http.StatusBadRequest,
			checkName:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			handler.GetUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkName && w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, user.Name, response["name"])
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed multiple users
	helpers.SeedUser(t, db, "User One", "UID001", "DID1")
	helpers.SeedUser(t, db, "User Two", "UID002", "DID2")
	helpers.SeedUser(t, db, "User Three", "UID003", "DID3")

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	tests := []struct {
		name        string
		params      string
		expectName  string
		expectCount int
	}{
		{"no filters", "", "User One", 3},
		{"search name", "?search=User+Two", "User Two", 1},
		{"limit 1", "?limit=1", "User One", 1},
		{"offset 2", "?offset=2&sort_by=created_at&order=asc", "User Three", 1},
		{"order desc", "?order=desc", "User Three", 3},
		{"sort_by created_at", "?sort_by=created_at", "User One", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/users" + tt.params
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			handler.ListUsers(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			var users []map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&users)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(users), 1)
			found := false
			for _, u := range users {
				if u["name"] == tt.expectName {
					found = true
				}
			}
			assert.True(t, found)
		})
	}
}

func TestExportUsersCSV(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed users
	helpers.SeedUser(t, db, "Export User 1", "UID001", "123456")
	helpers.SeedUser(t, db, "Export User 2", "UID002", "789012")
	helpers.SeedUser(t, db, "Export User 3", "UID003", "345678")

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	tests := []struct {
		name       string
		params     string
		expectName string
	}{
		{"no filters", "", "Export User 1"},
		{"search name", "?search=Export+User+2", "Export User 2"},
		{"limit 1", "?limit=1", "Export User 1"},
		{"offset 2", "?offset=2", "Export User 3"},
		{"order desc", "?order=desc", "Export User 3"},
		{"sort_by created_at", "?sort_by=created_at", "Export User 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/users/export" + tt.params
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			handler.ExportUsersCSV(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
			assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
			assert.Contains(t, w.Body.String(), "Name,RFID_UID,DiscordID")
			assert.Contains(t, w.Body.String(), tt.expectName)
		})
	}
}

func TestImportUsersCSV(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	// Create CSV content
	csvContent := `Name,RFID_UID,DiscordID
Import User 1,IMPORT001,111111
Import User 2,IMPORT002,222222`

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "users.csv")
	require.NoError(t, err)
	_, err = part.Write([]byte(csvContent))
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/users/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.ImportUsersCSV(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, float64(2), result["imported"])
	assert.Equal(t, float64(0), result["failed"])
}

// Edge case tests for user handlers

func TestImportUsersCSV_MalformedCSV(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	csvContent := `WrongHeader,AnotherWrong
Value1,Value2`

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "users.csv")
	require.NoError(t, err)
	_, err = part.Write([]byte(csvContent))
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/users/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.ImportUsersCSV(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImportUsersCSV_NoFile(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/users/import", nil)
	w := httptest.NewRecorder()

	handler.ImportUsersCSV(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	user := helpers.SeedUser(t, db, "Original Name", "RFID999", "discord_orig")

	updateReq := dto.UpdateUserRequest{
		Name:      "Updated Name",
		RFIDUID:   "RFID999_NEW",
		DiscordID: "discord_updated",
	}
	body, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", response["name"])
	assert.Equal(t, "discord_updated", response["discord_id"])

	updated, err := userRepo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "RFID999_NEW", updated.RFIDUID)
}

func TestUpdateUser_InvalidJSON(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.UpdateUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	user := helpers.SeedUser(t, db, "ToDelete", "RFID888", "discord_del")

	req := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	_, err := userRepo.FindByID(user.ID)
	assert.Error(t, err)
}

func TestListUsers_InvalidQueryParams(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	tests := []struct {
		name   string
		params string
	}{
		{"invalid limit", "?limit=abc"},
		{"negative limit", "?limit=-5"},
		{"invalid offset", "?offset=xyz"},
		{"negative offset", "?offset=-10"},
		{"invalid order", "?order=invalid"},
		{"invalid sort_by", "?sort_by=invalid_field"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/users"+tt.params, nil)
			w := httptest.NewRecorder()
			handler.ListUsers(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestUpdateUser_InvalidUserID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	updateReq := dto.UpdateUserRequest{Name: "Test"}
	body, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPut, "/api/users/abc", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUser_InvalidUserID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	handler := httptransport.NewUserHandler(userSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/invalid", nil)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

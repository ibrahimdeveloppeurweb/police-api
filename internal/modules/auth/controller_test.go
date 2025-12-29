package auth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/crypto"
	"police-trafic-api-frontend-aligned/internal/infrastructure/database"
	"police-trafic-api-frontend-aligned/internal/infrastructure/jwt"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestController creates a controller with all dependencies for testing
func setupTestController() *Controller {
	logger := zap.NewNop()

	// JWT service with test config
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-controller-testing",
			AccessExpiration: 15 * time.Minute,
		},
	}
	jwtService := jwt.NewJWTService(cfg, logger)

	// Crypto service
	cryptoService := crypto.NewPasswordService(logger)

	// Mock database
	mockDB := &database.MockDB{}

	// Auth service
	authService := NewService(logger, mockDB, jwtService, cryptoService)

	// Controller
	return NewController(authService, logger)
}

// Helper to create Echo context for testing
func createTestContext(method, path string, body interface{}) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()

	var reqBody bytes.Buffer
	if body != nil {
		json.NewEncoder(&reqBody).Encode(body)
	}

	req := httptest.NewRequest(method, path, &reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, rec
}

// Helper to add auth header to context
func addAuthHeader(c echo.Context, token string) {
	c.Request().Header.Set("Authorization", "Bearer "+token)
}

func TestAuthController_Login(t *testing.T) {
	controller := setupTestController()

	tests := []struct {
		name           string
		request        LoginRequest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "valid login",
			request: LoginRequest{
				Matricule: "12345",
				Password:  "any_password",
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				data := response["data"].(map[string]interface{})
				assert.NotEmpty(t, data["token"])

				user := data["user"].(map[string]interface{})
				assert.Equal(t, "12345", user["matricule"])
				assert.Equal(t, "Dupont", user["nom"])
				assert.Equal(t, "agent", user["role"])
			},
		},
		{
			name: "invalid matricule",
			request: LoginRequest{
				Matricule: "99999",
				Password:  "any_password",
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response["message"], "invalid credentials")
			},
		},
		{
			name: "empty password",
			request: LoginRequest{
				Matricule: "12345",
				Password:  "",
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, strings.ToLower(response["message"].(string)), "validation")
			},
		},
		{
			name: "missing matricule",
			request: LoginRequest{
				Password: "any_password",
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, strings.ToLower(response["message"].(string)), "validation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestContext("POST", "/auth/login", tt.request)

			err := controller.Login(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			tt.checkResponse(t, rec)
		})
	}
}

func TestAuthController_Register(t *testing.T) {
	controller := setupTestController()

	tests := []struct {
		name           string
		request        RegisterRequest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "valid registration",
			request: RegisterRequest{
				Matricule: "99999",
				Password:  "securePassword123",
				Nom:       "Test",
				Prenom:    "User",
				Email:     "test@police.gouv.fr",
				Role:      "agent",
			},
			expectedStatus: 201,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, "User registered successfully", response["message"])

				data := response["data"].(map[string]interface{})
				assert.Equal(t, "99999", data["matricule"])
				assert.Equal(t, "Test", data["nom"])
				assert.Equal(t, "test@police.gouv.fr", data["email"])
				assert.Equal(t, "agent", data["role"])
				assert.Equal(t, true, data["active"])
			},
		},
		{
			name: "invalid email",
			request: RegisterRequest{
				Matricule: "88888",
				Password:  "securePassword123",
				Nom:       "Test",
				Prenom:    "User",
				Email:     "invalid-email",
				Role:      "agent",
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, strings.ToLower(response["message"].(string)), "validation")
			},
		},
		{
			name: "invalid role",
			request: RegisterRequest{
				Matricule: "77777",
				Password:  "securePassword123",
				Nom:       "Test",
				Prenom:    "User",
				Email:     "test2@police.gouv.fr",
				Role:      "invalid_role",
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, strings.ToLower(response["message"].(string)), "validation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestContext("POST", "/auth/register", tt.request)

			err := controller.Register(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			tt.checkResponse(t, rec)
		})
	}
}

func TestAuthController_GetCurrentUser(t *testing.T) {
	controller := setupTestController()

	// First login to get a valid token
	loginReq := LoginRequest{
		Matricule: "12345",
		Password:  "any_password",
	}

	c, rec := createTestContext("POST", "/auth/login", loginReq)
	err := controller.Login(c)
	require.NoError(t, err)

	var loginResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &loginResp)
	require.NoError(t, err)

	data := loginResp["data"].(map[string]interface{})
	validToken := data["token"].(string)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid token",
			token:          validToken,
			expectedStatus: 200,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				data := response["data"].(map[string]interface{})
				assert.Equal(t, "12345", data["matricule"])
				assert.Equal(t, "Dupont", data["nom"])
				assert.Equal(t, "agent", data["role"])
			},
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: 401,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, strings.ToLower(response["message"].(string)), "unauthorized")
			},
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: 401,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, strings.ToLower(response["message"].(string)), "required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestContext("GET", "/auth/me", nil)

			if tt.token != "" {
				addAuthHeader(c, tt.token)
			}

			err := controller.GetCurrentUser(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			tt.checkResponse(t, rec)
		})
	}
}

func TestAuthController_RefreshToken(t *testing.T) {
	controller := setupTestController()

	// First login to get a valid token
	loginReq := LoginRequest{
		Matricule: "67890", // Admin user
		Password:  "any_password",
	}

	c, rec := createTestContext("POST", "/auth/login", loginReq)
	err := controller.Login(c)
	require.NoError(t, err)

	var loginResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &loginResp)
	require.NoError(t, err)

	data := loginResp["data"].(map[string]interface{})
	validToken := data["token"].(string)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid token refresh",
			token:          validToken,
			expectedStatus: 200,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				data := response["data"].(map[string]interface{})
				assert.NotEmpty(t, data["token"])

				user := data["user"].(map[string]interface{})
				assert.Equal(t, "67890", user["matricule"])
				assert.Equal(t, "admin", user["role"])
			},
		},
		{
			name:           "invalid token refresh",
			token:          "invalid-token",
			expectedStatus: 401,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, strings.ToLower(response["message"].(string)), "unauthorized")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestContext("POST", "/auth/refresh", nil)

			if tt.token != "" {
				addAuthHeader(c, tt.token)
			}

			err := controller.RefreshToken(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			tt.checkResponse(t, rec)
		})
	}
}

func TestAuthController_Logout(t *testing.T) {
	controller := setupTestController()

	// First login to get a valid token
	loginReq := LoginRequest{
		Matricule: "11111", // Supervisor user
		Password:  "any_password",
	}

	c, rec := createTestContext("POST", "/auth/login", loginReq)
	err := controller.Login(c)
	require.NoError(t, err)

	var loginResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &loginResp)
	require.NoError(t, err)

	data := loginResp["data"].(map[string]interface{})
	validToken := data["token"].(string)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid token logout",
			token:          validToken,
			expectedStatus: 200,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, "Logout successful", response["message"])
			},
		},
		{
			name:           "invalid token logout",
			token:          "invalid-token",
			expectedStatus: 401,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, strings.ToLower(response["message"].(string)), "unauthorized")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestContext("POST", "/auth/logout", nil)

			if tt.token != "" {
				addAuthHeader(c, tt.token)
			}

			err := controller.Logout(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			tt.checkResponse(t, rec)
		})
	}
}

func TestAuthController_CompleteFlow(t *testing.T) {
	controller := setupTestController()

	// Step 1: Register a new user
	registerReq := RegisterRequest{
		Matricule: "55555",
		Password:  "testPassword123",
		Nom:       "Flow",
		Prenom:    "Test",
		Email:     "flow@police.gouv.fr",
		Role:      "agent",
	}

	c, rec := createTestContext("POST", "/auth/register", registerReq)
	err := controller.Register(c)
	require.NoError(t, err)
	assert.Equal(t, 201, rec.Code)

	// Step 2: Login with the new user
	loginReq := LoginRequest{
		Matricule: registerReq.Matricule,
		Password:  registerReq.Password,
	}

	c, rec = createTestContext("POST", "/auth/login", loginReq)
	err = controller.Login(c)
	require.NoError(t, err)
	assert.Equal(t, 200, rec.Code)

	var loginResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &loginResp)
	require.NoError(t, err)

	data := loginResp["data"].(map[string]interface{})
	token := data["token"].(string)

	// Step 3: Get current user info
	c, rec = createTestContext("GET", "/auth/me", nil)
	addAuthHeader(c, token)
	err = controller.GetCurrentUser(c)
	require.NoError(t, err)
	assert.Equal(t, 200, rec.Code)

	var userResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &userResp)
	require.NoError(t, err)

	userData := userResp["data"].(map[string]interface{})
	assert.Equal(t, registerReq.Matricule, userData["matricule"])

	// Step 4: Refresh token
	c, rec = createTestContext("POST", "/auth/refresh", nil)
	addAuthHeader(c, token)
	err = controller.RefreshToken(c)
	require.NoError(t, err)
	assert.Equal(t, 200, rec.Code)

	var refreshResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &refreshResp)
	require.NoError(t, err)

	refreshData := refreshResp["data"].(map[string]interface{})
	newToken := refreshData["token"].(string)

	// Step 5: Logout with new token
	c, rec = createTestContext("POST", "/auth/logout", nil)
	addAuthHeader(c, newToken)
	err = controller.Logout(c)
	require.NoError(t, err)
	assert.Equal(t, 200, rec.Code)
}

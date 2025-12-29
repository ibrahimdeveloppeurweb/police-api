package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/modules/auth"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// TestServer provides utilities for integration testing
type TestServer struct {
	App    *echo.Echo
	Server *httptest.Server
	Config *config.Config
}

// TestAuthTokens holds authentication tokens for different roles
type TestAuthTokens struct {
	Admin      string
	Supervisor string
	Agent      string
}

// NewTestServer creates a new test server for integration testing
func NewTestServer(t *testing.T) *TestServer {
	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         "8080",
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-integration-testing",
			AccessExpiration: 30 * 60, // 30 minutes in seconds
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			DBName:   "test_db",
			User:     "test_user",
			Password: "test_password",
			SSLMode:  "disable",
		},
	}

	// For now, let's create a simpler test setup without full DI
	// We'll manually wire the dependencies needed for auth testing
	echoApp := echo.New()
	
	// We would need to manually setup routes here, but for now let's use a simpler approach
	// This is a placeholder - in real implementation, we'd setup the full app
	
	// Create test server
	server := httptest.NewServer(echoApp)
	
	return &TestServer{
		App:    echoApp,
		Server: server,
		Config: cfg,
	}
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// GetTestTokens generates authentication tokens for all roles for testing
func (ts *TestServer) GetTestTokens(t *testing.T) *TestAuthTokens {
	tokens := &TestAuthTokens{}
	
	// Test users for login
	testUsers := []struct {
		role      string
		matricule string
		password  string
		tokenPtr  *string
	}{
		{"admin", "67890", "any_password", &tokens.Admin},
		{"supervisor", "11111", "any_password", &tokens.Supervisor},
		{"agent", "12345", "any_password", &tokens.Agent},
	}

	for _, user := range testUsers {
		loginReq := auth.LoginRequest{
			Matricule: user.matricule,
			Password:  user.password,
		}

		resp, err := ts.PostJSON("/auth/login", loginReq)
		require.NoError(t, err, "Login should succeed for %s", user.role)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Login should return 200 for %s", user.role)

		var loginResp auth.LoginResponse
		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err, "Should decode login response for %s", user.role)
		require.NotEmpty(t, loginResp.Token, "Token should not be empty for %s", user.role)

		*user.tokenPtr = loginResp.Token
		resp.Body.Close()
	}

	return tokens
}

// PostJSON sends a POST request with JSON payload
func (ts *TestServer) PostJSON(path string, payload interface{}) (*http.Response, error) {
	return ts.RequestWithJSON("POST", path, payload, "")
}

// PostJSONWithAuth sends a POST request with JSON payload and authorization header
func (ts *TestServer) PostJSONWithAuth(path string, payload interface{}, token string) (*http.Response, error) {
	return ts.RequestWithJSON("POST", path, payload, token)
}

// GetWithAuth sends a GET request with authorization header
func (ts *TestServer) GetWithAuth(path string, token string) (*http.Response, error) {
	return ts.Request("GET", path, nil, token)
}

// Get sends a GET request
func (ts *TestServer) Get(path string) (*http.Response, error) {
	return ts.Request("GET", path, nil, "")
}

// PutJSONWithAuth sends a PUT request with JSON payload and authorization header
func (ts *TestServer) PutJSONWithAuth(path string, payload interface{}, token string) (*http.Response, error) {
	return ts.RequestWithJSON("PUT", path, payload, token)
}

// DeleteWithAuth sends a DELETE request with authorization header
func (ts *TestServer) DeleteWithAuth(path string, token string) (*http.Response, error) {
	return ts.Request("DELETE", path, nil, token)
}

// RequestWithJSON sends a request with JSON payload
func (ts *TestServer) RequestWithJSON(method, path string, payload interface{}, token string) (*http.Response, error) {
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return nil, err
		}
	}

	return ts.Request(method, path, &body, token)
}

// Request sends a HTTP request to the test server
func (ts *TestServer) Request(method, path string, body *bytes.Buffer, token string) (*http.Response, error) {
	url := ts.Server.URL + path
	
	var bodyReader *bytes.Buffer
	if body != nil {
		bodyReader = body
	} else {
		bodyReader = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{}
	return client.Do(req)
}

// AssertErrorResponse checks that the response is an error with expected status and message
func (ts *TestServer) AssertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedMessageContains string) {
	require.Equal(t, expectedStatus, resp.StatusCode)
	
	var errorResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)
	
	message, exists := errorResp["message"].(string)
	require.True(t, exists, "Error response should have message field")
	
	if expectedMessageContains != "" {
		require.Contains(t, message, expectedMessageContains)
	}
}

// AssertSuccessResponse checks that the response is successful and optionally validates data
func (ts *TestServer) AssertSuccessResponse(t *testing.T, resp *http.Response, expectedStatus int) map[string]interface{} {
	require.Equal(t, expectedStatus, resp.StatusCode)
	
	var successResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&successResp)
	require.NoError(t, err)
	
	return successResp
}

// AssertJSONResponse decodes response into provided struct
func (ts *TestServer) AssertJSONResponse(t *testing.T, resp *http.Response, expectedStatus int, target interface{}) {
	require.Equal(t, expectedStatus, resp.StatusCode)
	
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err)
}
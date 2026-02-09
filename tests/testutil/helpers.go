package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// authResponseData matches the API response for register/login
type authResponseData struct {
	Data struct {
		User  interface{} `json:"user"`
		Token string      `json:"token"`
	} `json:"data"`
}

// Register sends POST /api/auth/register and returns the token. Returns error on non-2xx or parse failure.
func Register(router *gin.Engine, name, email, password string) (token string, status int, err error) {
	body := map[string]string{"name": name, "email": email, "password": password}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	status = w.Code
	if status < 200 || status >= 300 {
		return "", status, nil
	}
	var resp authResponseData
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		return "", status, err
	}
	return resp.Data.Token, status, nil
}

// MustRegister calls Register and fails the test if registration fails. Returns the JWT token.
func MustRegister(t *testing.T, router *gin.Engine, name, email, password string) string {
	t.Helper()
	token, status, err := Register(router, name, email, password)
	if err != nil {
		t.Fatalf("register parse response: %v", err)
	}
	if status != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", status)
	}
	if token == "" {
		t.Fatal("register: empty token")
	}
	return token
}

// Login sends POST /api/auth/login and returns the token. Returns error on parse failure.
func Login(router *gin.Engine, email, password string) (token string, status int, err error) {
	body := map[string]string{"email": email, "password": password}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	status = w.Code
	if status < 200 || status >= 300 {
		return "", status, nil
	}
	var resp authResponseData
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		return "", status, err
	}
	return resp.Data.Token, status, nil
}

// MustLogin calls Login and fails the test if login fails. Returns the JWT token.
func MustLogin(t *testing.T, router *gin.Engine, email, password string) string {
	t.Helper()
	token, status, err := Login(router, email, password)
	if err != nil {
		t.Fatalf("login parse response: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", status)
	}
	if token == "" {
		t.Fatal("login: empty token")
	}
	return token
}

// Request performs an HTTP request against the router and returns the response.
// Use for custom requests (e.g. with Authorization header for protected routes).
func Request(router *gin.Engine, method, path string, body []byte, authToken string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

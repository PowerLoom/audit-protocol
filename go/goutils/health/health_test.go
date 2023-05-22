package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	// Create a new HTTP request to the "/health" endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Create a response recorder to record the HTTP response
	res := httptest.NewRecorder()

	// Call the HealthCheck handler function directly
	HealthCheckHandler().ServeHTTP(res, req)

	// Check the response status code
	if res.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, res.Code)
	}
}

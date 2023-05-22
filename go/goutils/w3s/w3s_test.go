package w3storage

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
)

func TestUploadToW3s_Success(t *testing.T) {
	expectedCID := "mockCID"

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Assert the request
		assert.Equal(t, "/upload", req.URL.Path)
		assert.Equal(t, "Bearer mockToken", req.Header.Get("Authorization"))

		// Marshal the expected response
		response := struct {
			CID string `json:"CID"`
		}{
			CID: expectedCID,
		}
		respBody, _ := json.Marshal(response)

		// Send the response
		rw.WriteHeader(http.StatusOK)
		rw.Write(respBody)
	}))

	defer server.Close()

	settingsObj := &settings.SettingsObj{
		HttpClient: &settings.HTTPClient{
			MaxIdleConns:        1,
			MaxConnsPerHost:     1,
			MaxIdleConnsPerHost: 1,
			IdleConnTimeout:     60,
		},
		Web3Storage: &settings.Web3Storage{
			URL:             server.URL,
			UploadURLSuffix: "/upload",
			APIToken:        "mockToken",
		},
	}

	w := &W3S{
		limiter:           rate.NewLimiter(1, 1),
		settingsObj:       settingsObj,
		defaultHTTPClient: httpclient.GetDefaultHTTPClient(settingsObj),
	}

	// Invoke the function
	cid, err := w.UploadToW3s(map[string]interface{}{"mock": "data"})

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, expectedCID, cid)
}

func TestUploadToW3s_Error(t *testing.T) {
	expectedErrorMessage := "mock error message"

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Assert the request
		assert.Equal(t, "/upload", req.URL.Path)
		assert.Equal(t, "Bearer mockToken", req.Header.Get("Authorization"))

		// Marshal the expected error response
		response := struct {
			Message string `json:"message"`
		}{
			Message: expectedErrorMessage,
		}
		respBody, _ := json.Marshal(response)

		// Send the response with an error status code
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(respBody)
	}))
	defer server.Close()

	settingsObj := &settings.SettingsObj{
		HttpClient: &settings.HTTPClient{
			MaxIdleConns:        1,
			MaxConnsPerHost:     1,
			MaxIdleConnsPerHost: 1,
			IdleConnTimeout:     60,
		},
		Web3Storage: &settings.Web3Storage{
			URL:             server.URL,
			UploadURLSuffix: "/upload",
			APIToken:        "mockToken",
		},
	}

	w := &W3S{
		limiter:           rate.NewLimiter(1, 1),
		settingsObj:       settingsObj,
		defaultHTTPClient: httpclient.GetDefaultHTTPClient(settingsObj),
	}

	// Invoke the function
	cid, err := w.UploadToW3s(map[string]interface{}{"mock": "data"})

	// Assert the results
	assert.Error(t, err)
	assert.Equal(t, expectedErrorMessage, err.Error())
	assert.Empty(t, cid)
}

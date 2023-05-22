package transactions

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/assert"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/settings"
)

func TestSendSignatureToRelayer(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "/endpoint", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read the request body
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		// Unmarshal the request body into a struct
		var rb datamodel.RelayerRequest
		err = json.Unmarshal(body, &rb)
		assert.NoError(t, err)

		// Assert the expected values in the request
		expectedPayload := &datamodel.SnapshotRelayerPayload{
			ProjectID:   "your_project_id",
			SnapshotCID: "your_snapshot_cid",
			EpochID:     1,
			Request: map[string]interface{}{
				"deadline": math.NewHexOrDecimal256(2),
			},
			Signature: "your_signature",
		}
		assert.Equal(t, expectedPayload.ProjectID, rb.ProjectID)
		assert.Equal(t, expectedPayload.SnapshotCID, rb.SnapshotCID)
		assert.Equal(t, expectedPayload.EpochID, rb.EpochID)
		assert.Equal(t, expectedPayload.Request["deadline"], math.NewHexOrDecimal256(int64(rb.Request.Deadline)))
		assert.Equal(t, "0x"+expectedPayload.Signature, rb.Signature)

		// Write a mock response
		w.WriteHeader(http.StatusOK)
	}))

	defer mockServer.Close()

	// Create a test instance of the TxManager with mock settings
	endpoint := "/endpoint"
	txManager := &TxManager{
		settingsObj: &settings.SettingsObj{
			Relayer: &settings.Relayer{
				Host:     &mockServer.URL,
				Endpoint: &endpoint,
			},
			HttpClient: &settings.HTTPClient{
				MaxIdleConns:        1,
				MaxConnsPerHost:     1,
				MaxIdleConnsPerHost: 1,
				IdleConnTimeout:     60,
			},
		},
	}

	// Create a test payload
	payload := &datamodel.SnapshotRelayerPayload{
		ProjectID:   "your_project_id",
		SnapshotCID: "your_snapshot_cid",
		EpochID:     1,
		Request: map[string]interface{}{
			"deadline": math.NewHexOrDecimal256(2),
		},
		Signature: "your_signature",
	}

	// Call the function
	err := txManager.SendSignatureToRelayer(payload)

	// Check the result
	assert.NoError(t, err)
}

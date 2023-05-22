package reporting

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"audit-protocol/goutils/settings"
)

func TestIssueReporter_ReportOnSlack(t *testing.T) {
	// Create a test settings object
	settingsObj := &settings.SettingsObj{
		HttpClient: &settings.HTTPClient{
			MaxIdleConns:        1,
			MaxConnsPerHost:     1,
			MaxIdleConnsPerHost: 1,
			IdleConnTimeout:     60,
		},
		Reporting: &settings.Reporting{
			SlackWebhookURL: "http://example.com/slack-webhook",
		},
	}

	// Create an instance of the IssueReporter with the mock dependencies
	reporter := InitIssueReporter(settingsObj)

	// Create a sample issue payload
	issue := []byte(`{"text":"Sample issue"}`)

	// Set up a mock HTTP server to simulate the response from the Slack webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/slack-webhook", r.URL.String())

		// Simulate a successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Success"}`))
	}))
	defer server.Close()

	// Update the Slack webhook URL in the settings object to the mock server URL
	settingsObj.Reporting.SlackWebhookURL = server.URL + "/slack-webhook"

	// Call the function under test
	reporter.ReportOnSlack(issue)
}

func TestIssueReporter_ReportToOffchainConsensus(t *testing.T) {
	// Create a test settings object
	settingsObj := &settings.SettingsObj{
		HttpClient: &settings.HTTPClient{
			MaxIdleConns:        1,
			MaxConnsPerHost:     1,
			MaxIdleConnsPerHost: 1,
			IdleConnTimeout:     60,
		},
		Reporting: &settings.Reporting{},
	}

	// Create an instance of the IssueReporter with the mock HTTP client and settings object
	reporter := InitIssueReporter(settingsObj)

	// Create a sample issue payload
	issue := []byte(`{"text":"Sample issue"}`)

	// Set up a mock HTTP server to simulate the response from the offchain consensus endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/offchain-consensus", r.URL.String())

		// Simulate a successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Success"}`))
	}))
	defer server.Close()

	// Update the OffchainConsensusIssueEndpoint in the settings object to the mock server URL
	settingsObj.Reporting.OffchainConsensusIssueEndpoint = server.URL + "/offchain-consensus"

	// Call the function under test
	reporter.ReportToOffchainConsensus(issue)
}

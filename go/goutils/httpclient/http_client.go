package httpclient

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"audit-protocol/goutils/settings"
)

// GetDefaultHTTPClient returns a retryablehttp.Client with default values
// use this method for default http client needs for specific settings create custom method
func GetDefaultHTTPClient(settingsObj *settings.SettingsObj) *retryablehttp.Client {
	transport := &http.Transport{
		MaxIdleConns:        settingsObj.HttpClient.MaxIdleConns,
		MaxConnsPerHost:     settingsObj.HttpClient.MaxConnsPerHost,
		MaxIdleConnsPerHost: settingsObj.HttpClient.MaxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(settingsObj.HttpClient.IdleConnTimeout) * time.Second,
		DisableCompression:  true,
	}

	rawHTTPClient := &http.Client{
		Transport: transport,
	}

	retryableHTTPClient := retryablehttp.NewClient()
	retryableHTTPClient.RetryMax = 5
	retryableHTTPClient.HTTPClient = rawHTTPClient

	return retryableHTTPClient
}

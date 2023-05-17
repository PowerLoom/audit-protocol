package httpclient

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/settings"
)

// GetDefaultHTTPClient returns a retryablehttp.Client with default values
// use this method for default http client needs for specific settings create custom method
func GetDefaultHTTPClient() *retryablehttp.Client {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke settings object")
	}

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

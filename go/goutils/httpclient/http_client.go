package httpclient

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

// GetDefaultHTTPClient returns a retryablehttp.Client with default values
// use this method for default http client needs for specific settings create custom method
func GetDefaultHTTPClient() *retryablehttp.Client {
	transport := &http.Transport{
		MaxIdleConns:        5,
		MaxConnsPerHost:     5,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     0,
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

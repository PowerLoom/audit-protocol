package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/dnscache"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/settings"
)

var dnsResolver *dnscache.Resolver

func init() {
	dnsResolver = &dnscache.Resolver{}

	go func() {
		clearUnused := true
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for range t.C {
			dnsResolver.Refresh(clearUnused)
		}
	}()
}

// GetDefaultHTTPClient returns a retryablehttp.Client with default values
// use this method for default http client needs for specific settings create custom method
func GetDefaultHTTPClient(timeout int) *retryablehttp.Client {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke settings object")
	}

	rawHTTPClient := &http.Client{
		Transport: getDefaultHTTPTransport(settingsObj),
		Timeout:   time.Duration(timeout) * time.Second,
	}

	retryableHTTPClient := retryablehttp.NewClient()
	retryableHTTPClient.RetryMax = 5
	retryableHTTPClient.HTTPClient = rawHTTPClient
	retryableHTTPClient.Logger = log.StandardLogger()

	return retryableHTTPClient
}

func GetIPFSWriteHTTPClient(settingsObj *settings.SettingsObj) *http.Client {
	if settingsObj.IpfsConfig.WriterAuthConfig.ProjectApiKey == "" || settingsObj.IpfsConfig.WriterAuthConfig.ProjectApiSecret == "" {
		return GetDefaultHTTPClient(settingsObj.IpfsConfig.Timeout).HTTPClient
	}

	return &http.Client{
		Transport: ipfsAuthTransport{
			RoundTripper:  getDefaultHTTPTransport(settingsObj),
			ProjectKey:    settingsObj.IpfsConfig.WriterAuthConfig.ProjectApiKey,
			ProjectSecret: settingsObj.IpfsConfig.WriterAuthConfig.ProjectApiSecret,
		},
		Timeout: time.Duration(settingsObj.IpfsConfig.Timeout) * time.Second,
	}
}

func GetIPFSReadHTTPClient(settingsObj *settings.SettingsObj) *http.Client {
	if settingsObj.IpfsConfig.ReaderAuthConfig.ProjectApiKey == "" || settingsObj.IpfsConfig.ReaderAuthConfig.ProjectApiSecret == "" {
		return GetDefaultHTTPClient(settingsObj.IpfsConfig.Timeout).HTTPClient
	}

	return &http.Client{
		Transport: ipfsAuthTransport{
			RoundTripper:  getDefaultHTTPTransport(settingsObj),
			ProjectKey:    settingsObj.IpfsConfig.ReaderAuthConfig.ProjectApiKey,
			ProjectSecret: settingsObj.IpfsConfig.ReaderAuthConfig.ProjectApiSecret,
		},
		Timeout: time.Duration(settingsObj.IpfsConfig.Timeout) * time.Second,
	}
}

func getDefaultHTTPTransport(settingsObj *settings.SettingsObj) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			ips, err := dnsResolver.LookupHost(ctx, host)
			if err != nil {
				return nil, err
			}

			for _, ip := range ips {
				var dialer net.Dialer
				conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
				if err == nil {
					break
				}
			}

			return
		},
		MaxIdleConns:        settingsObj.HttpClient.MaxIdleConns,
		MaxConnsPerHost:     settingsObj.HttpClient.MaxConnsPerHost,
		MaxIdleConnsPerHost: settingsObj.HttpClient.MaxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(settingsObj.HttpClient.IdleConnTimeout) * time.Second,
	}
}

// ipfsAuthTransport decorates each request with a basic auth header.
type ipfsAuthTransport struct {
	http.RoundTripper
	ProjectKey    string
	ProjectSecret string
}

func (t ipfsAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.ProjectKey, t.ProjectSecret)

	return t.RoundTripper.RoundTrip(r)
}

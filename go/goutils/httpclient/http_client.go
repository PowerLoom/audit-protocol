package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"audit-protocol/goutils/settings"
	"github.com/rs/dnscache"
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
func GetDefaultHTTPClient(settingsObj *settings.SettingsObj) *retryablehttp.Client {
	transport := &http.Transport{
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

	rawHTTPClient := &http.Client{
		Transport: transport,
	}

	retryableHTTPClient := retryablehttp.NewClient()
	retryableHTTPClient.RetryMax = 5
	retryableHTTPClient.HTTPClient = rawHTTPClient

	return retryableHTTPClient
}

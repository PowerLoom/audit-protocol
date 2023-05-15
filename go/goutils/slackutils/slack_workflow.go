package slackutils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"
	"golang.org/x/time/rate"

	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityError    Severity = "ERROR"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
)

type Report map[string]interface{}

type SlackNotifyReq struct {
	Data          map[string]interface{} `json:"errorDetails"`
	IssueSeverity string                 `json:"severity"`
	Service       string                 `json:"service"`
}

type SlackResp struct {
	Error            string `json:"error"`
	Ok               bool   `json:"ok"`
	ResponseMetadata struct {
		Messages []string `json:"messages"`
	} `json:"response_metadata"`
}

type SlackClient struct {
	slackClient *retryablehttp.Client
	rateLimiter *rate.Limiter
	url         string
}

func InitSlackWorkFlowClient() *SlackClient {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.WithError(err).Fatalf("failed to invoke settings object")
	}

	client := &SlackClient{
		slackClient: httpclient.GetDefaultHTTPClient(),
		rateLimiter: rate.NewLimiter(1, 1),
		url:         settingsObj.SlackWebhookURL,
	}

	err = gi.Inject(client)
	if err != nil {
		log.WithError(err).Fatalf("failed to inject slack client")
	}

	return client
}

func (s *SlackClient) NotifySlackWorkflow(reportData Report, severity Severity, service string) {
	_ = s.rateLimiter.Wait(context.Background())

	reqURL := s.url

	if reqURL == "" {
		return
	}

	slackReq := &SlackNotifyReq{
		Data:          reportData,
		IssueSeverity: string(severity),
		Service:       service,
	}

	data, err := json.Marshal(slackReq)
	if err != nil {
		log.WithError(err).Error("failed to marshal slack request body")

		return
	}

	body, err := json.Marshal(map[string]interface{}{"text": string(data)})
	if err != nil {
		log.WithError(err).Error("CRITICAL: failed to marshal slack request body")

		return
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		log.WithError(err).Error("failed to create request to slack webhook url")

		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")

	log.Debugf("sending req with params to slack webhook url")

	res, err := s.slackClient.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to send request to slack webhook url")

		return
	}

	defer res.Body.Close()

	if err != nil {
		log.WithError(err).Error("failed to read response body from slack webhook url")

		return
	}

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Error("failed to read response body from slack webhook url")
	}

	if res.StatusCode == http.StatusOK {
		log.WithField("resp", string(resp)).Debug("Successfully sent request to Slack Webhook")

		return
	}

	log.WithField("resp", string(resp)).Info("response from slack")
}

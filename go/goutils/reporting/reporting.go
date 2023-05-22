package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
)

type IssueType string

const (
	PayloadCommitInternalIssue      IssueType = "PAYLOAD_COMMIT_INTERNAL_ISSUE" // generic issue type for internal errors
	MissedSnapshotIssue             IssueType = "MISSED_SNAPSHOT"               // when a snapshot was missed
	SubmittedIncorrectSnapshotIssue IssueType = "SUBMITTED_INCORRECT_SNAPSHOT"  // when a snapshot was submitted but it was incorrect
)

type Service interface {
	Report(issueType IssueType, projectID string, epochID string, extra map[string]interface{})
}

type IssueReporter struct {
	httpClient       *retryablehttp.Client
	slackRateLimiter *rate.Limiter
	settingsObj      *settings.SettingsObj
}

func InitIssueReporter(settingsObj *settings.SettingsObj) *IssueReporter {
	client := &IssueReporter{
		httpClient:       httpclient.GetDefaultHTTPClient(settingsObj),
		slackRateLimiter: rate.NewLimiter(1, 1),
		settingsObj:      settingsObj,
	}

	return client
}

func (i *IssueReporter) Report(issueType IssueType, projectID string, epochID string, extra map[string]interface{}) {
	extraData, err := json.Marshal(extra)
	if err != nil {
		log.WithError(err).Error("failed to marshal extra data")
	}

	issue := &datamodel.SnapshotterIssue{
		InstanceID:      i.settingsObj.InstanceId,
		IssueType:       string(issueType),
		ProjectID:       projectID,
		EpochID:         epochID,
		TimeOfReporting: strconv.FormatInt(time.Now().Unix(), 10),
		Extra:           string(extraData),
	}

	log.WithField("issue", issue).Debug("reporting issue")

	issueBytes, err := json.Marshal(issue)
	if err != nil {
		log.WithError(err).Error("failed to json marshal issue")

		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		i.ReportOnSlack(issueBytes)
	}()

	go func() {
		defer wg.Done()
		i.ReportToOffchainConsensus(issueBytes)
	}()

	wg.Wait()
}

func (i *IssueReporter) ReportOnSlack(issue []byte) {
	if i.settingsObj.Reporting.SlackWebhookURL == "" {
		return
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, i.settingsObj.Reporting.SlackWebhookURL, bytes.NewBuffer(issue))
	if err != nil {
		log.WithError(err).Error("failed to create request to slack webhook url")

		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")

	log.Debugf("sending issue to slack")

	err = i.slackRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithError(err).Error("failed to wait for slack rate limiter")

		return
	}

	res, err := i.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to send request to slack webhook")

		return
	}

	defer res.Body.Close()

	if err != nil {
		log.WithError(err).Error("failed to read response body from slack webhook")

		return
	}

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Error("failed to read response body from slack webhook")
	}

	if res.StatusCode == http.StatusOK {
		log.WithField("resp", string(resp)).Debug("status ok response from slack webhook")

		return
	}

	log.WithField("resp", string(resp)).Info("response from slack webhook")
}

func (i *IssueReporter) ReportToOffchainConsensus(issue []byte) {
	req, err := retryablehttp.NewRequest(http.MethodPost, i.settingsObj.Reporting.OffchainConsensusIssueEndpoint, bytes.NewBuffer(issue))
	if err != nil {
		log.WithError(err).Error("failed to create request to slack webhook url")

		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")

	log.Debug("sending issue to offchain consensus")

	res, err := i.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to send request to offchain consensus")

		return
	}

	defer res.Body.Close()

	if err != nil {
		log.WithError(err).Error("failed to read response body from offchain consensus")

		return
	}

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Error("failed to read response body from offchain consensus")
	}

	if res.StatusCode == http.StatusOK {
		log.WithField("resp", string(resp)).Debug("status ok response from offchain consensus")

		return
	}

	log.WithField("resp", string(resp)).Info("response from offchain consensus")
}

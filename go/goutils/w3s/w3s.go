package w3storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
)

type Service interface {
	UploadToW3s(msg interface{}) (string, error)
}

type W3S struct {
	limiter           *rate.Limiter
	settingsObj       *settings.SettingsObj
	defaultHTTPClient *retryablehttp.Client
}

func InitW3S(settingsObj *settings.SettingsObj) Service {
	log.Debug("initializing web3.storage client")

	// Default values
	tps := rate.Limit(1) // 3 TPS
	burst := 1

	if settingsObj.Web3Storage.RateLimiter != nil {
		burst = settingsObj.Web3Storage.RateLimiter.Burst

		if settingsObj.Web3Storage.RateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.Web3Storage.RateLimiter.RequestsPerSec)
		}
	}

	log.Infof("rate Limit configured for web3.storage at %v TPS with a burst of %d", tps, burst)

	rateLimiter := rate.NewLimiter(tps, burst)

	w := &W3S{
		limiter:           rateLimiter,
		settingsObj:       settingsObj,
		defaultHTTPClient: httpclient.GetDefaultHTTPClient(settingsObj),
	}

	return w
}

func (w *W3S) UploadToW3s(msg interface{}) (string, error) {
	log.Debug("uploading payload commit message to web3 storage")

	reqURL := w.settingsObj.Web3Storage.URL + w.settingsObj.Web3Storage.UploadURLSuffix

	payloadCommit, err := json.Marshal(msg)
	if err != nil {
		log.WithError(err).Error("failed to marshal payload commit message")

		return "", err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(payloadCommit))
	if err != nil {
		log.WithError(err).Error("failed to create request to web3.storage")

		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+w.settingsObj.Web3Storage.APIToken)
	req.Header.Add("accept", "application/json")

	err = w.limiter.Wait(context.Background())
	if err != nil {
		log.Errorf("web3 storage rate limiter wait errored")

		return "", err
	}

	log.WithField("msg", string(payloadCommit)).Debug("sending request to web3.storage")

	res, err := w.defaultHTTPClient.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to send request to web3.storage")

		return "", err
	}

	defer res.Body.Close()

	web3resp := new(datamodel.Web3StoragePutResponse)

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Error("failed to read response body from web3.storage")

		return "", err
	}

	if res.StatusCode == http.StatusOK {
		err = json.Unmarshal(respBody, web3resp)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal response from web3.storage")

			return "", err
		}

		log.WithField("cid", web3resp.CID).Info("successfully uploaded payload commit to web3.storage")

		return web3resp.CID, nil
	}

	resp := new(datamodel.Web3StorageErrResponse)

	err = json.Unmarshal(respBody, resp)
	if err != nil {
		log.WithError(err).Error("failed to unmarshal error response from web3.storage")
	} else {
		log.WithField("payloadCommit", string(payloadCommit)).WithField("error", resp.Message).Error("web3.storage upload error")
	}

	return "", errors.New(resp.Message)
}

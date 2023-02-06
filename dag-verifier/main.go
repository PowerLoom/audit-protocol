package main

import (
	// "context"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"github.com/powerloom/audit-prototol-private/goutils/datamodel"
	"github.com/powerloom/audit-prototol-private/goutils/ipfsutils"
	"github.com/powerloom/audit-prototol-private/goutils/logger"
	"github.com/powerloom/audit-prototol-private/goutils/redisutils"
	"github.com/powerloom/audit-prototol-private/goutils/settings"
	"github.com/powerloom/audit-prototol-private/goutils/slackutils"
)

var ipfsClient ipfsutils.IpfsClient
var dagVerifier DagVerifier

var settingsObj *settings.SettingsObj

func main() {
	logger.InitLogger()
	settingsObj = settings.ParseSettings("../settings.json")
	dagVerifier.Initialize(settingsObj)
	var wg sync.WaitGroup

	http.HandleFunc("/reportIssue", IssueReportHandler)
	port := settingsObj.DagVerifierSettings.Port
	host := settingsObj.DagVerifierSettings.Host
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Starting HTTP server on port %d in a go routine.", port)
		http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	}()

	if settingsObj.DagVerifierSettings.PruningVerification {
		var pruningVerifier PruningVerifier
		pruningVerifier.Init(settingsObj)
		wg.Add(1)

		go func() {
			defer wg.Done()
			pruningVerifier.Run()
		}()
	}

	dagVerifier.Run()

	if settingsObj.DagVerifierSettings.PruningVerification {
		wg.Wait()
	}
}

func IssueReportHandler(w http.ResponseWriter, req *http.Request) {
	log.Infof("Received issue report %+v : ", *req)
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("Failed to read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Record issues in redis
	res := dagVerifier.redisClient.ZAdd(ctx, redisutils.REDIS_KEY_ISSUES_REPORTED, &redis.Z{Score: float64(time.Now().UnixMicro()),
		Member: reqBytes,
	})
	if res.Err() != nil {
		log.Errorf("Failed to add issue to redis due to error %+v", res.Err())
	}
	go func() {

		// Notify consensus layer
		ReportIssueToConsensus(reqBytes)

		var reqPayload datamodel.IssueReport

		err = json.Unmarshal(reqBytes, &reqPayload)
		if err != nil {
			log.Errorf("Error while parsing json body of issue report %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		//Notify on slack and report to consensus layer
		report, _ := json.MarshalIndent(reqPayload, "", "\t")
		slackutils.NotifySlackWorkflow(string(report), reqPayload.Severity, reqPayload.Service)
		//Prune issues older than 7 days
		pruneTillTime := time.Now().Add(-7 * 24 * 60 * 60 * time.Second).UnixMicro()
		dagVerifier.redisClient.ZRemRangeByScore(ctx, redisutils.REDIS_KEY_ISSUES_REPORTED, "0", fmt.Sprintf("%d", pruneTillTime))
	}()
	w.WriteHeader(http.StatusOK)
}

func ReportIssueToConsensus(reqBytes []byte) {
	reqURL := settingsObj.ConsensusConfig.ServiceURL + "/reportIssue"
	for retryCount := 0; ; {
		if retryCount == 3 {
			log.Errorf("failed to send issueReport to consensus service after max-retry")
			return
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(reqBytes))
		if err != nil {
			log.Fatalf("Failed to create new HTTP Req with URL %s due to error %+v",
				reqURL, err)
			return
		}
		req.Header.Add("accept", "application/json")
		log.Debugf("Sending issue report %+v to consensus service URL %s",
			req, reqURL)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			retryCount++
			log.Errorf("Failed to send request %+v towards consensus service URL %s due to error %+v.  Retrying %d",
				req, reqURL, err, retryCount)
			continue
		}
		defer res.Body.Close()
		respBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			retryCount++
			log.Errorf("Failed to read response body from consensus service with error %+v.",
				err, retryCount)
			break
		}
		if res.StatusCode == http.StatusOK {
			log.Infof("Reported issue to consensus layer.")
			break
		} else {
			retryCount++
			log.Errorf("Received Error response %+v from consensus service with statusCode %d and status : %s ",
				respBody, res.StatusCode, res.Status)
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
	}
}

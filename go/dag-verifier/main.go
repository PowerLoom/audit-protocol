package main

import (
	// "context"

	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/caching"
	"audit-protocol/dag-verifier/verifiers"
	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/slackutils"
)

func main() {
	logger.InitLogger()

	settingsObj := settings.ParseSettings()

	ipfsURL := settingsObj.IpfsConfig.ReaderURL
	if ipfsURL == "" {
		ipfsURL = settingsObj.IpfsConfig.URL
	}

	_ = redisutils.InitRedisClient(
		settingsObj.Redis.Host,
		settingsObj.Redis.Port,
		settingsObj.Redis.Db,
		settingsObj.DagVerifierSettings.RedisPoolSize,
		settingsObj.Redis.Password,
		-1,
	)

	_ = ipfsutils.InitClient(
		ipfsURL,
		settingsObj.DagVerifierSettings.Concurrency,
		settingsObj.DagVerifierSettings.IPFSRateLimiter,
		settingsObj.HttpClientTimeoutSecs,
	)

	slackutils.InitSlackWorkFlowClient(settingsObj.DagVerifierSettings.SlackNotifyURL)

	redisCache := caching.NewRedisCache()

	err := gi.Inject(redisCache)
	if err != nil {
		log.Panicln("failed to inject dependencies", err)
		return
	}

	_ = caching.InitDiskCache()

	_ = verifiers.InitDagVerifier()

	var wg sync.WaitGroup

	http.HandleFunc("/reportIssue", IssueReportHandler)
	http.HandleFunc("/dagBlocksInserted", DagBlocksInsertedHandler)
	port := settingsObj.DagVerifierSettings.Port
	hostPort := net.JoinHostPort(settingsObj.DagVerifierSettings.Host, strconv.Itoa(port))

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Infof("starting HTTP server on port %d in a go routine.", port)

		err := http.ListenAndServe(hostPort, nil)
		if err != nil {
			log.Error("failed to start HTTP server", err)
			// if server fails to start then exit
			os.Exit(1)
		}
	}()

	if settingsObj.DagVerifierSettings.PruningVerification {
		pruningVerifier, err := verifiers.InitPruningVerifier()
		if err != nil {
			log.Error("failed to initialize the pruning verifier", err)

			return
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			pruningVerifier.Run()
		}()
	}

	wg.Wait()
}

// DagBlocksInsertedHandler handles the dag blocks inserted event callback
func DagBlocksInsertedHandler(w http.ResponseWriter, r *http.Request) {
	// read body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read request body for dag blocks inserted event")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// parse body
	dagBlocksInserted := new(verifiers.DagBlocksInsertedReq)

	err = json.Unmarshal(reqBody, dagBlocksInserted)
	if err != nil {
		log.Errorf("failed to parse request body for dag blocks inserted event")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	dagVerifier, err := gi.Invoke[*verifiers.DagVerifier]()
	if err != nil {
		log.Errorf("failed to get dag verifier instance")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	go dagVerifier.Run(dagBlocksInserted)

	w.WriteHeader(http.StatusOK)
}

// IssueReportHandler handles the issue report http request
func IssueReportHandler(w http.ResponseWriter, req *http.Request) {
	log.Infof("Received issue report %+v : ", *req)
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("Failed to read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	issueReport := new(datamodel.IssueReport)
	err = json.Unmarshal(reqBytes, issueReport)
	if err != nil {
		log.Errorf("Failed to parse issue report")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	redisCache, err := gi.Invoke[*caching.RedisCache]()
	if err != nil {
		log.Errorf("Failed to get redis cache instance")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	//Record issues in redis
	err = redisCache.StoreReportedIssues(context.Background(), issueReport)
	if err != nil {
		log.Errorf("Failed to add issue to redis due to error %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go func(issueReport *datamodel.IssueReport) {
		// Notify consensus layer
		ReportIssueToConsensus(reqBytes)

		//Notify on slack and report to consensus layer
		report, _ := json.MarshalIndent(issueReport, "", "\t")
		err = slackutils.NotifySlackWorkflow(string(report), issueReport.Severity, issueReport.Service)
		if err != nil {
			log.Errorf("Failed to notify slack due to error %+v", err)
		}

		// Prune issues older than 7 days.
		pruneTillTime := time.Now().Add(-7 * 24 * 60 * 60 * time.Second).UnixMicro()

		err = redisCache.RemoveOlderReportedIssues(context.Background(), pruneTillTime)
		if err != nil {
			log.Errorf("Failed to prune older sissues from cache due to error %+v", err)
		}
	}(issueReport)

	w.WriteHeader(http.StatusOK)
}

func ReportIssueToConsensus(reqBytes []byte) {
	settingsObj, _ := gi.Invoke[*settings.SettingsObj]()

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

		respBody, err := io.ReadAll(res.Body)
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

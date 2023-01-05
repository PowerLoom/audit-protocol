package main

import (
	// "context"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/powerloom/goutils/logger"
	"github.com/powerloom/goutils/settings"
	"github.com/powerloom/goutils/slackutils"
)

var ipfsClient IpfsClient

func main() {
	logger.InitLogger()
	settingsObj := settings.ParseSettings("../settings.json")
	var dagVerifier DagVerifier
	dagVerifier.Initialize(settingsObj)
	var wg sync.WaitGroup

	http.HandleFunc("/reportIssue", IssueReportHandler)
	port := settingsObj.DagVerifierSettings.IssueReporterPort
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Starting HTTP server on port %d in a go routine.", port)
		http.ListenAndServe(fmt.Sprint(":", port), nil)
	}()

	//For now just using settings to determine this in case multiple namespaces are being run.
	//In future dag verifier would also work with multiple namespaces, this can be removed once implemented.
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
	log.Infof("Received block height confirm callback %+v : ", *req)
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("Failed to read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var reqPayload IssueReport

	err = json.Unmarshal(reqBytes, &reqPayload)
	if err != nil {
		log.Errorf("Error while parsing json body of issue report %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//Notify on slack and report to consensus layer
	report, _ := json.MarshalIndent(reqPayload, "", "\t")
	slackutils.NotifySlackWorkflow(string(report), reqPayload.Severity, reqPayload.Service)
	//TODO: Notify consensus layer
	w.WriteHeader(http.StatusOK)
}

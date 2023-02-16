package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

/* type TestPayload struct {
	Field1 string `json:f1`
	Field2 string `json:f2`
}

type PayloadCommit struct {
	ProjectId            string       `json:"projectId"`
	CommitId             string       `json:"commitId"`
	Payload              *TestPayload `json:"payload,omitempty""`
	XYZ                  string       `json:"xyz"`
	TentativeBlockHeight int64        `json:"tentativeBlockHeight"`
	SnapshotCID          string       `json:"snapshotCID"`
	ApiKeyHash           string       `json:"apiKeyHash"`
	Resubmitted          bool         `json:"resubmitted"`
	ResubmissionBlock    int          `json:"resubmissionBlock"` // corresponds to lastTouchedBlock in PendingTransaction model
	Web3Storage          bool         `json:"web3Storage"`
	SkipAnchorProof      bool         `json:"skipAnchorProof"`
}
type AuditContractCommitResp struct {
	Success bool                          `json:"success"`
	Data    []AuditContractCommitRespData `json:"data"`
	Error   AuditContractErrResp          `json:"error"`
}
type AuditContractCommitRespData struct {
	TxHash string `json:"txHash"`
}

type AuditContractErrResp struct {
	Message string `json:"message"`
	Error   struct {
		Message string `json:"message"`
		Details struct {
			BriefMessage string `json:"briefMessage"`
			FullMessage  string `json:"fullMessage"`
			Data         []struct {
				Contract       string          `json:"contract"`
				Method         string          `json:"method"`
				Params         json.RawMessage `json:"params"`
				EncodingErrors struct {
					APIKeyHash string `json:"apiKeyHash"`
				} `json:"encodingErrors"`
			} `json:"data"`
		} `json:"details"`
	} `json:"error"`
}
*/
/*
func writerHandler(w http.ResponseWriter, req *http.Request) {
	log.Infof("Received http request %+v : ", *req)
	var resp AuditContractCommitResp
	resp.Success = true
	var respData AuditContractCommitRespData
	respData.TxHash = fmt.Sprint("tokenHash:", rand.Int63())
	resp.Data = append(resp.Data, respData)

	res, _ := json.Marshal(resp)
	//w.WriteHeader(http.StatusInternalServerError)

	length, err := w.Write(res)
	if err != nil {
		log.Error("Failed to write res with err:", err)
	}
	log.Infof("Sent response %+v of length %d", resp, length)
}

func WebhookListener(w http.ResponseWriter, req *http.Request) {
	log.Infof("Received http request for callback %+v : ", *req)
	reqBytes, _ := ioutil.ReadAll(req.Body)
	var reqParams map[string]json.RawMessage
	json.Unmarshal(reqBytes, &reqParams)
	log.Infof("Received http request body for callback %+v : ", reqParams)
	var resp AuditContractCommitResp
	resp.Success = true

	res, _ := json.Marshal(resp)
	//w.WriteHeader(http.StatusInternalServerError)

	length, err := w.Write(res)
	if err != nil {
		log.Error("Failed to write res with err:", err)
	}
	log.Infof("Sent response %+v for webhook callback of length %d", resp, length)
}
*/

func InitLogger() {
	log.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.TraceLevel,
			log.InfoLevel,
			log.DebugLevel,
		},
	})
	if len(os.Args) < 2 {
		fmt.Println("Pass loglevel as an argument if you don't want default(INFO) to be set.")
		fmt.Println("Values to be passed for logLevel: ERROR(2),INFO(4),DEBUG(5)")
		log.SetLevel(log.DebugLevel)
	} else {
		logLevel, err := strconv.ParseUint(os.Args[1], 10, 32)
		if err != nil || logLevel > 6 {
			log.SetLevel(log.DebugLevel) //TODO: Change default level to error
		} else {
			//TODO: Need to come up with approach to dynamically update logLevel.
			log.SetLevel(log.Level(logLevel))
		}
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

type SubmitSnapshotResponse struct {
	Status               string `json:"status"`
	DelayedSubmission    bool   `json:"delayedSubmission"`
	FinalizedSnapshotCID string `json:"finalizedSnapshotCID"`
}

const SNAPSHOT_CONSENSUS_STATUS_ACCEPTED string = "ACCEPTED"
const SNAPSHOT_CONSENSUS_STATUS_FINALIZED string = "FINALIZED"

type SubmitSnapshotRequest struct {
	Epoch       int64  `json:"epoch"`
	ProjectID   string `json:"projectID"`
	InstanceID  string `json:"instanceID"`
	SnapshotCID string `json:"snapshotCID"`
}

func SimulateConsensusServer() {
	InitLogger()
	http.HandleFunc("/submitSnapshot", DetermineConsensus)
	http.HandleFunc("/checkForSnapshotConfirmation", DetermineConsensus)
	port := 9030
	log.SetLevel(log.DebugLevel)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Starting HTTP server on port %d in a go routine.", port)
		http.ListenAndServe(fmt.Sprint(":", port), nil)
	}()
	wg.Wait()
}

func FNV32a(text string) uint32 {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(text))
	return algorithm.Sum32()
}

func DetermineConsensus(w http.ResponseWriter, req *http.Request) {
	var snapshotReq SubmitSnapshotRequest
	bytes, _ := ioutil.ReadAll(req.Body)
	json.Unmarshal(bytes, &snapshotReq)
	log.Infof("Received http request for URL %s %+v : ", req.URL, snapshotReq)
	hash := FNV32a(snapshotReq.ProjectID)
	status := SNAPSHOT_CONSENSUS_STATUS_FINALIZED

	if strings.Contains(req.URL.Path, "submitSnapshot") && hash%2 == 0 {
		status = SNAPSHOT_CONSENSUS_STATUS_ACCEPTED
	}
	resp := SubmitSnapshotResponse{Status: status, DelayedSubmission: false, FinalizedSnapshotCID: snapshotReq.SnapshotCID}
	respBytes, _ := json.Marshal(resp)
	log.Infof("Responding 200 OK with %+v", resp)
	w.Write(respBytes)
}

func main() {
	SimulateConsensusServer()
	/* conn, err := GetConn("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	numProjects := 10
	var tentativeBlockHeights [50]int64
	http.HandleFunc("/writer", writerHandler)
	http.HandleFunc("/", WebhookListener)

	port := 8888
	rand.Seed(time.Now().Unix())
	log.SetLevel(log.DebugLevel)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Starting HTTP server on port %d in a go routine.", port)
		http.ListenAndServe(fmt.Sprint(":", port), nil)
	}()

	rand.Seed(time.Now().Unix())
	for j := 0; j < 5; j++ {
		log.Infof("Sending %d burst to rabbitmq.", j)
		for i := 0; i < numProjects; i++ {

			tentativeBlockHeights[i] += 1
			var pc PayloadCommit
			pc.CommitId = "testCommitId" + fmt.Sprint(rand.Int())
			pc.ProjectId = "testProjectId" + fmt.Sprint(i)
			pc.TentativeBlockHeight = tentativeBlockHeights[i]
			if i%2 == 0 {
				pc.Web3Storage = true
				pc.SkipAnchorProof = false
			} else {
				pc.SkipAnchorProof = true
			}
			if j%2 == 0 { //simulate resubmissions intermittently
				//Simulate a resubmission.
				// Create a cid manually by specifying the 'prefix' parameters
				pref := cid.Prefix{
					Version:  1,
					Codec:    (uint64)(mc.Raw),
					MhType:   mh.SHA2_256,
					MhLength: -1, // default length
				}
				pc.Resubmitted = true
				pc.ResubmissionBlock = j
				pcBytes, _ := json.Marshal(pc)
				// And then feed it some data
				c, err := pref.Sum(pcBytes)
				if err != nil {
					log.Error("Failed to add bytes to CID")
				}

				fmt.Println("Created CID: ", c)
				pc.SnapshotCID = c.String()
				pc.Payload = nil
			} else {
				pc.Resubmitted = false
				pc.Payload = &TestPayload{Field1: "Hello", Field2: "Go Payload Commit service." + fmt.Sprint(rand.Int())}
				pc.XYZ = "Test which should not be included in payload"
			}
			bytes, _ := json.Marshal(pc)
			log.Info("payloadCommit.Payload", pc.Payload)
			log.Infof("Sending %+v to rabbitmq.", pc)
			//fmt.Println("Sending bytes", bytes)
			err := conn.Publish("commit-payloads", bytes)
			if err != nil {
				log.Error("Failed to publish msg to rabbitmq.", err)
			}
			//break
		}
		//break
		time.Sleep(5 * time.Second)
	}
	wg.Wait() */
}

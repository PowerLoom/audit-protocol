package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	cid "github.com/ipfs/go-cid"
	mc "github.com/multiformats/go-multicodec"
	mh "github.com/multiformats/go-multihash"

	log "github.com/sirupsen/logrus"
)

type TestPayload struct {
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

func main() {
	conn, err := GetConn("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	numProjects := 10
	var tentativeBlockHeights [50]int64
	http.HandleFunc("/writer", writerHandler)
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
	wg.Wait()
}

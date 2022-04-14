package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

type TestPayload struct {
	Field1 string `json:f1`
	Field2 string `json:f2`
}

type PayloadCommit struct {
	ProjectId            string      `json:"projectId"`
	CommitId             string      `json:"commitId"`
	Payload              TestPayload `json:"payload"`
	TentativeBlockHeight int64       `json:"tentativeBlockHeight"`
}

func main() {
	conn, err := GetConn("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	numProjects := 50
	var tentativeBlockHeights [50]int64
	for j := 0; ; j++ {
		log.Infof("Sending %d burst to rabbitmq.", j)
		for i := 0; i < numProjects; i++ {
			tentativeBlockHeights[i] += 1
			var pc PayloadCommit
			pc.CommitId = "testCommitId" + fmt.Sprint(rand.Int())
			pc.ProjectId = "testProjectId" + fmt.Sprint(i)
			pc.TentativeBlockHeight = tentativeBlockHeights[i]
			pc.Payload = TestPayload{Field1: "Hello", Field2: "Go Payload Commit service." + fmt.Sprint(rand.Int())}
			bytes, _ := json.Marshal(pc)
			log.Debugf("Sending +%v to rabbitmq.", pc)
			//fmt.Println("Sending bytes", bytes)
			err := conn.Publish("commit-payloads", bytes)
			if err != nil {
				log.Error("Failed to publish msg to rabbitmq.")
			}
		}
		time.Sleep(5 * 1000000000)
	}
}

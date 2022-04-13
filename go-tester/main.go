package main

import (
	"encoding/json"
	"fmt"
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
	var pc PayloadCommit
	pc.CommitId = "testCommitId"
	pc.ProjectId = "testProjectId"
	pc.TentativeBlockHeight = 1
	pc.Payload = TestPayload{Field1: "Hello", Field2: "Go Payload Commit service."}
	bytes, err := json.Marshal(pc)
	fmt.Println("Sending +%v to rabbitmq.", pc)
	fmt.Println("Sending bytes", bytes)
	conn.Publish("commit-payloads", bytes)
}

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/settings"
)

var settingsObj = &settings.SettingsObj{
	LocalCachePath: "/tmp",
	Pruning: &settings.Pruning{
		IPFSPinningMaxAge: 1,
		LocalDiskMaxAge:   1,
	},
}

type MockIPFSService struct{}

func (m MockIPFSService) UploadSnapshotToIPFS(snapshot *datamodel.PayloadCommitMessage) error {
	// TODO implement me
	panic("implement me")
}

func (m MockIPFSService) GetSnapshotFromIPFS(snapshotCID string, outputPath string) error {
	// TODO implement me
	panic("implement me")
}

func (m MockIPFSService) Unpin(cid string) error {
	return nil
}

var _ ipfsutils.Service = &MockIPFSService{}

func Test_prune(t *testing.T) {
	err := prune(settingsObj, new(MockIPFSService))
	assert.Equal(t, err, nil)

	// read invalid directory
	settingsObj.LocalCachePath = "/tmp/invalid"
	err = prune(settingsObj, new(MockIPFSService))
	assert.NotEqual(t, err, nil)

	// create tmp directory
	settingsObj.LocalCachePath = "/tmp/pruning_test"

	dirPath := filepath.Join(settingsObj.LocalCachePath, "project1", "snapshots")
	filePath := filepath.Join(dirPath, "bafybeihdwdcefgh4dqkjv67uzcmw7ojee6xedzdetojuzjevtenxquvyku.json")

	_ = os.MkdirAll(dirPath, os.ModePerm)
	defer os.RemoveAll(dirPath)

	_, err = os.Create(filePath)
	assert.Equal(t, err, nil)

	err = os.Chtimes(filePath, time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -2))
	assert.Equal(t, err, nil)

	err = prune(settingsObj, new(MockIPFSService))
	assert.Equal(t, err, nil)

	// check if file is removed
	_, err = os.Stat(filePath)
	assert.NotEqual(t, err, nil)
}

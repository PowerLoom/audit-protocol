package pruning

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"audit-protocol/goutils/mock"
	"audit-protocol/goutils/settings"
)

var settingsObj = &settings.SettingsObj{
	LocalCachePath: "/tmp",
	Pruning: &settings.Pruning{
		IPFSPinningMaxAge: 1,
		LocalDiskMaxAge:   1,
	},
}

func Test_prune(t *testing.T) {
	ipfsServiceMock := new(mock.IPFSServiceMock)

	ipfsServiceMock.UnpinMock = func(cid string) error {
		return nil
	}

	err := Prune(settingsObj, ipfsServiceMock)
	assert.Equal(t, err, nil)

	// read invalid directory
	settingsObj.LocalCachePath = "/tmp/invalid"
	err = Prune(settingsObj, ipfsServiceMock)
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

	err = Prune(settingsObj, ipfsServiceMock)
	assert.Equal(t, err, nil)

	// check if file is removed
	_, err = os.Stat(filePath)
	assert.NotEqual(t, err, nil)
}

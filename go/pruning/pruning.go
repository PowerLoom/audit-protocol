package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/settings"
)

func main() {
	logger.InitLogger()

	settingsObj := settings.ParseSettings()

	ipfsClient := ipfsutils.InitClient(
		settingsObj.IpfsConfig.URL,
		settingsObj.IpfsConfig.IPFSRateLimiter,
		settingsObj.IpfsConfig.Timeout,
	)

	cronRunner := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger),
	))

	// Run every 7days
	cronId, err := cronRunner.AddFunc(settingsObj.Pruning.CronFrequency, func() {
		prune(settingsObj, ipfsClient)
	})
	if err != nil {
		log.WithError(err).Fatal("failed to add pruning cron job")
	}

	log.WithField("cronId", cronId).Info("added pruning cron job")

	cronRunner.Start()

	// block forever
	select {}
}

func prune(settingsObj *settings.SettingsObj, client *ipfsutils.IpfsClient) {
	log.Debug("ipfs unpinning started")

	// get all files from local disk cache.
	dirEntries, err := os.ReadDir(settingsObj.LocalCachePath)
	if err != nil {
		log.WithError(err).Error("failed to read local cache dir")
	}

	cidsToUnpin := make([]string, 0)

	// read all the files from the directory.
	for _, dirEntry := range dirEntries {
		// dirEntry should be a directory (looking for project directories).
		if !dirEntry.IsDir() {
			continue
		}

		projectDir, err := os.ReadDir(filepath.Join(settingsObj.LocalCachePath, dirEntry.Name()))
		if err != nil {
			log.WithError(err).Error("failed to read project dir")

			continue
		}

		// read all the files from the directory
		for _, projectDirEntry := range projectDir {
			// dirEntry should be a file <cid.json>
			if projectDirEntry.IsDir() {
				continue
			}

			ext := path.Ext(projectDirEntry.Name())
			if !strings.EqualFold(ext, ".json") {
				continue
			}

			// check if file is older than 7 days, skip if not
			info, err := projectDirEntry.Info()
			if err != nil {
				continue
			}

			modTime := info.ModTime()
			if modTime.AddDate(0, 0, settingsObj.Pruning.MaxAge).After(time.Now()) {
				continue
			}

			// get the file name without extension
			fileName := strings.TrimSuffix(projectDirEntry.Name(), ext)

			// check if fileName is valid cid format
			cid, err := cid.Parse(fileName)
			if err != nil {
				continue
			}

			cidsToUnpin = append(cidsToUnpin, cid.String())
		}
	}

	if len(cidsToUnpin) == 0 {
		log.Debug("nothing to unpin")

		return
	}

	// unpin all the cids.
	wg := sync.WaitGroup{}
	for _, c := range cidsToUnpin {
		wg.Add(1)
		log.Debug("unpinning cid: ", c)

		go func(cidToUnpin string) {
			defer wg.Done()

			err = client.Unpin(cidToUnpin)
			if err != nil {
				log.WithField("cid", cidToUnpin).WithError(err).Error("failed to unpin cid")

				return
			}
		}(c)
	}

	wg.Wait()

	log.Info("ipfs unpinning completed")
}

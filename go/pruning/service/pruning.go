package pruning

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/remeh/sizedwaitgroup"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/settings"
)

const ServiceName = "pruning"

func Prune(settingsObj *settings.SettingsObj, ipfsService ipfsutils.Service) error {
	log.Debug("pruning started")

	// get all files from local disk cache.
	dirEntries, err := os.ReadDir(settingsObj.LocalCachePath)
	if err != nil {
		log.WithError(err).Error("failed to read local cache dir")

		return err
	}

	cidsToUnpin := make([]string, 0)
	filesToRemove := make([]string, 0)

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

		// check for snapshots directory.
		for _, projectDirEntry := range projectDir {
			// check if directory name is "snapshots"
			if !strings.EqualFold(projectDirEntry.Name(), "snapshots") {
				continue
			}

			// dirEntry should be a file <cid.json>
			if !projectDirEntry.IsDir() {
				continue
			}

			snapshotsDir, err := os.ReadDir(filepath.Join(settingsObj.LocalCachePath, dirEntry.Name(), projectDirEntry.Name()))
			if err != nil {
				log.WithError(err).Error("failed to read snapshots dir")

				continue
			}

			for _, snapshotsDirEntry := range snapshotsDir {
				ext := path.Ext(snapshotsDirEntry.Name())
				if !strings.EqualFold(ext, ".json") {
					continue
				}

				// check if file is older than 7 days, skip if not
				info, err := snapshotsDirEntry.Info()
				if err != nil {
					continue
				}

				modTime := info.ModTime()
				if modTime.AddDate(0, 0, settingsObj.Pruning.IPFSPinningMaxAge).Before(time.Now()) {
					// get the file name without extension
					fileName := strings.TrimSuffix(snapshotsDirEntry.Name(), ext)

					// check if fileName is valid cid format
					cid, err := cid.Parse(fileName)
					if err != nil {
						continue
					}

					cidsToUnpin = append(cidsToUnpin, cid.String())
				}

				// add file to filesToRemove
				if modTime.AddDate(0, 0, settingsObj.Pruning.LocalDiskMaxAge).Before(time.Now()) {
					absFilePath := filepath.Join(settingsObj.LocalCachePath, dirEntry.Name(), projectDirEntry.Name(), snapshotsDirEntry.Name())

					filesToRemove = append(filesToRemove, absFilePath)
				}
			}
		}
	}

	if len(cidsToUnpin) == 0 {
		log.Debug("no cids to unpin")
	}

	if len(filesToRemove) == 0 {
		log.Debug("no files to remove")
	}

	// unpin all the cids.
	wg := new(sync.WaitGroup)

	wg.Add(1)

	go func() {
		defer wg.Done()

		swg := sizedwaitgroup.New(settingsObj.Concurrency)

		for _, c := range cidsToUnpin {
			swg.Add()
			log.Debug("unpinning cid: ", c)

			go func(cidToUnpin string) {
				defer swg.Done()

				err = ipfsService.Unpin(cidToUnpin)
				if err != nil {
					log.WithField("cid", cidToUnpin).WithError(err).Error("failed to unpin cid")

					return
				}
			}(c)
		}

		swg.Wait()
	}()

	wg.Add(1)

	go func() {
		swg := sizedwaitgroup.New(settingsObj.Concurrency)

		defer wg.Done()

		for _, file := range filesToRemove {
			swg.Add()
			log.Debug("removing file: ", file)

			go func(fileToRemove string) {
				defer swg.Done()

				err = os.Remove(fileToRemove)
				if err != nil {
					log.WithField("file", fileToRemove).WithError(err).Error("failed to remove file")

					return
				}
			}(file)
		}

		swg.Wait()
	}()

	wg.Wait()

	log.Info("ipfs unpinning completed")
	log.Info("local disk cleanup completed")

	return nil
}

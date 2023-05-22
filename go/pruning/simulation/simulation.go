package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/smartcontract"
	pruning "audit-protocol/pruning/service"
)

type State struct {
	SyncedTillEpochId     int64                    `json:"synced_till_epoch_id"`
	ProjectSpecificStates map[string]*ProjectState `json:"project_specific_states"`
}

type ProjectState struct {
	FirstEpochID  int64                  `json:"first_epoch_id"`
	FinalizedCIDs map[string]interface{} `json:"finalized_cids"`
}

func main() {
	logger.InitLogger()

	settingsObj := settings.ParseSettings()

	// state export's file path
	log.Info("reading state file")

	contents, err := os.ReadFile("../state.json")
	if err != nil {
		log.Fatal(err)
	}

	state := new(State)

	err = json.Unmarshal(contents, state)
	if err != nil {
		log.Fatal(err)
	}

	contractApi := smartcontract.InitContractAPI()

	log.Info("getting projects from contract")
	projects, err := contractApi.GetProjects(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}

	projectToCIDsMapping := make(map[string][]cid.Cid)

	log.Info("creating project to cids mapping")
	for _, project := range projects {
		projectState := state.ProjectSpecificStates[project]
		if projectState == nil {
			continue
		}

		for _, cidData := range projectState.FinalizedCIDs {
			// check the type of cid interface
			switch c := cidData.(type) {
			case string:
				parsedCid, err := cid.Parse(c)
				if err != nil {
					continue
				}

				projectToCIDsMapping[project] = append(projectToCIDsMapping[project], parsedCid)
			}
		}
	}

	ipfsClient := ipfsutils.InitClient(settingsObj.IpfsConfig.URL, settingsObj.IpfsConfig.IPFSRateLimiter, settingsObj.IpfsConfig.Timeout)

	err = os.MkdirAll(settingsObj.LocalCachePath, 0755)
	if err != nil {
		log.WithError(err).Fatal("failed to create cache dir")
	}

	wg := new(sync.WaitGroup)

	for project, cids := range projectToCIDsMapping {
		// storing snapshots locally
		dirPath := filepath.Join(settingsObj.LocalCachePath, project, "snapshots")

		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			log.WithError(err).Error("failed to create project cache dir", project)

			continue
		}

		log.Info("creating project cache files")

		for _, c := range cids {
			go func(cid cid.Cid) {
				wg.Add(1)
				defer wg.Done()

				filePath := filepath.Join(dirPath, c.String()+".json")

				err = os.WriteFile(filePath, []byte(c.String()), 0644)
				if err != nil {
					log.WithError(err).Error("failed to write cid to file", c.String())

					return
				}

				// change mod time of file to simulate as old files
				err = os.Chtimes(filePath, time.Now().AddDate(0, 0, -(settingsObj.Pruning.LocalDiskMaxAge+2)), time.Now().AddDate(0, 0, -(settingsObj.Pruning.LocalDiskMaxAge+2)))
				if err != nil {
					log.WithError(err).Error("failed to change file mod time", c.String())

					return
				}
			}(c)
		}
	}

	wg.Wait()

	// start pruning
	pruning.Prune(settingsObj, ipfsClient)

	log.Info("pruning completed")

	log.Info("checking if all files are pruned")

	for project, cids := range projectToCIDsMapping {
		dirPath := filepath.Join(settingsObj.LocalCachePath, project, "snapshots")

		for _, c := range cids {
			filePath := filepath.Join(dirPath, c.String()+".json")

			_, err = os.Stat(filePath)
			if err == nil {
				log.Error("file not pruned ", filePath)
			}
		}
	}

	log.Info("simulation completed")
}

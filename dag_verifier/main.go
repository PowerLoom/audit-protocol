package main

import (
	// "context"

	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/powerloom/goutils/logger"
	"github.com/powerloom/goutils/settings"
)

var ipfsClient IpfsClient
var pairContractAddresses []string

func main() {
	logger.InitLogger()
	settingsObj := settings.ParseSettings("../settings.json")
	var pairContractAddress string
	if len(os.Args) == 3 {
		pairContractAddress = os.Args[2]
	}
	PopulatePairContractList(pairContractAddress, "../static/cached_pair_addresses.json")
	var dagVerifier DagVerifier
	dagVerifier.Initialize(settingsObj, &pairContractAddresses)
	dagVerifier.Run()
}

func PopulatePairContractList(pairContractAddr string, pairContractListFile string) {
	if pairContractAddr != "" {
		log.Info("Skipping reading contract addresses from json.\nConsidering only passed pairContractaddress:", pairContractAddr)
		pairContractAddresses = make([]string, 1)
		pairContractAddresses[0] = pairContractAddr
		return
	}

	log.Info("Reading contracts:", pairContractListFile)
	data, err := os.ReadFile(pairContractListFile)
	if err != nil {
		log.Error("Cannot read the file:", err)
		panic(err)
	}

	log.Debug("Contracts json data is", string(data))
	err = json.Unmarshal(data, &pairContractAddresses)
	if err != nil {
		log.Error("Cannot unmarshal the pair-contracts json ", err)
		panic(err)
	}
}

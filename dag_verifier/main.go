package main

import (
	// "context"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

var ipfsClient IpfsClient
var pairContractAddresses []string

func main() {
	initLogger()
	settingsObj := ParseSettings("../settings.json")
	var pairContractAddress string
	if len(os.Args) == 3 {
		pairContractAddress = os.Args[2]
	}
	PopulatePairContractList(pairContractAddress, "../static/cached_pair_addresses.json")
	var dagVerifier DagVerifier
	dagVerifier.Initialize(settingsObj, &pairContractAddresses)
	dagVerifier.Run()
}

func initLogger() {
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
		log.SetLevel(log.InfoLevel)
	} else {
		logLevel, err := strconv.ParseUint(os.Args[1], 10, 32)
		if err != nil || logLevel > 6 {
			log.SetLevel(log.InfoLevel)
		} else {
			//TODO: Need to come up with approach to dynamically update logLevel.
			log.SetLevel(log.Level(logLevel))
		}
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
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

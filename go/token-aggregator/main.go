package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"

    log "github.com/sirupsen/logrus"
    "github.com/swagftw/gi"

    "audit-protocol/caching"
    "audit-protocol/goutils/logger"
    "audit-protocol/goutils/redisutils"
    "audit-protocol/goutils/settings"
    "audit-protocol/token-aggregator/models"
    "audit-protocol/token-aggregator/service"
)

func main() {
    // first read config settings
    logger.InitLogger()

    settingsObj := settings.ParseSettings()

    _ = redisutils.InitRedisClient(
        settingsObj.Redis.Host,
        settingsObj.Redis.Port,
        settingsObj.Redis.Db,
        settingsObj.DagVerifierSettings.RedisPoolSize,
        settingsObj.Redis.Password,
        -1,
    )

    caching.NewRedisCache()

    tokenAggregatorService := service.InitTokenAggService()

    var pairContractAddress string
    if len(os.Args) == 3 {
        pairContractAddress = os.Args[2]
    }

    // starting service
    go tokenAggregatorService.Run(pairContractAddress)

    // starting http server
    http.HandleFunc("/block_height_confirm_callback", blockHeightConfirmCallback)
    port := settingsObj.TokenAggregatorSettings.Port

    go func() {
        log.Infof("Starting HTTP server on port %d in a go routine.", port)
        err := http.ListenAndServe(fmt.Sprint(":", port), nil)
        if err != nil {
            log.WithError(err).Error("error starting http server")
            os.Exit(1)
        }
    }()

    select {}
}

func blockHeightConfirmCallback(w http.ResponseWriter, req *http.Request) {
    log.Infof("received block height confirm callback")

    reqBytes, _ := io.ReadAll(req.Body)
    reqPayload := new(models.BlockHeightConfirmationPayload)

    err := json.Unmarshal(reqBytes, reqPayload)
    if err != nil {
        log.WithError(err).Error("failed to unmarshal block height confirm callback payload")

        return
    }

    log.WithField("payload", reqPayload).Info("block height confirm callback payload")

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    resp := make(map[string]string)
    resp["message"] = "callback received"

    jsonResp, err := json.Marshal(resp)
    if err != nil {
        log.WithError(err).Fatalf("failed to marshal response to callback confirmation")
    }

    _, err = w.Write(jsonResp)
    if err != nil {
        log.WithError(err).Error("error while writing response to callback confirmation")
    }

    go func() {
        tokenAggService, err := gi.Invoke[*service.TokenAggregator]()
        if err != nil {
            log.WithError(err).Error("error while invoking token aggregator service")

            return
        }

        err = tokenAggService.FetchAndUpdateStatusOfOlderSnapshots(reqPayload.ProjectID)
        if err != nil {
            log.WithError(err).Error("error while fetching and updating status of older snapshots")

            return
        }
    }()
}

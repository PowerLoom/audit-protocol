package main

import (
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/slackutils"
	taskmgr "audit-protocol/goutils/taskmgr/rabbitmq"
	ps "audit-protocol/pruning-archival/service"
	"audit-protocol/pruning-archival/worker"

	log "github.com/sirupsen/logrus"
)

func main() {
	logger.InitLogger()

	// load settings
	settingsObj := settings.ParseSettings()
	settingsObj.PruningServiceSettings = settingsObj.GetDefaultPruneConfig()

	ipfsURL := settingsObj.IpfsConfig.ReaderURL
	if ipfsURL == "" {
		ipfsURL = settingsObj.IpfsConfig.URL
	}

	ipfsClient := ipfsutils.InitClient(
		ipfsURL,
		settingsObj.PruningServiceSettings.Concurrency,
		settingsObj.PruningServiceSettings.IPFSRateLimiter,
		settingsObj.PruningServiceSettings.IpfsTimeout,
	)

	redisClient := redisutils.InitRedisClient(
		settingsObj.Redis.Host,
		settingsObj.Redis.Port,
		settingsObj.Redis.Db,
		settingsObj.DagVerifierSettings.RedisPoolSize,
		settingsObj.Redis.Password,
		0,
	)

	ps.InitPruningService(settingsObj, redisClient, ipfsClient)

	slackutils.InitSlackWorkFlowClient(settingsObj.DagVerifierSettings.SlackNotifyURL)

	// this rabbitmq task manager
	taskmgr.NewRabbitmqTaskMgr()

	// init worker and start consuming tasks
	wkr := worker.NewWorker()

	for {
		if err := wkr.ConsumeTask(); err != nil {
			log.WithError(err).Error("error while consuming task")
		}
	}
}

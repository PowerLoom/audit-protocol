package main

import (
	log "github.com/sirupsen/logrus"

	"audit-protocol/caching"
	"audit-protocol/goutils/health"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/smartcontract"
	taskmgr "audit-protocol/goutils/taskmgr/rabbitmq"
	w3storage "audit-protocol/goutils/w3s"
	"audit-protocol/payload-commit/service"
	"audit-protocol/payload-commit/worker"
)

func main() {
	logger.InitLogger()
	settingsObj := settings.ParseSettings()

	ipfsutils.InitClient(
		settingsObj.IpfsConfig.URL,
		settingsObj.IpfsConfig.IPFSRateLimiter,
		settingsObj.IpfsConfig.Timeout,
	)

	redisClient := redisutils.InitRedisClient(
		settingsObj.Redis.Host,
		settingsObj.Redis.Port,
		settingsObj.Redis.Db,
		settingsObj.Redis.PoolSize,
		settingsObj.Redis.Password,
		-1,
	)

	caching.NewRedisCache()
	smartcontract.InitContractAPI()
	taskmgr.NewRabbitmqTaskMgr()
	w3storage.InitW3S()
	caching.InitDiskCache()

	service.InitPayloadCommitService()

	mqWorker := worker.NewWorker()

	// health check is non-blocking health check http listener
	health.HealthCheck()

	defer func() {
		mqWorker.ShutdownWorker()
		err := redisClient.Close()
		if err != nil {
			log.WithError(err).Error("error while closing redis client")
		}
	}()

	for {
		err := mqWorker.ConsumeTask()
		if err != nil {
			log.WithError(err).Error("error while consuming task, starting again")
		}
	}
}

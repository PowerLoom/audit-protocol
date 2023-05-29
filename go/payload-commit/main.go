package main

import (
	log "github.com/sirupsen/logrus"

	"audit-protocol/caching"
	"audit-protocol/goutils/health"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/reporting"
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

	ipfsutils.InitClient(settingsObj)

	readRedisClient := redisutils.InitReaderRedisClient(
		settingsObj.RedisReader.Host,
		settingsObj.RedisReader.Port,
		settingsObj.RedisReader.Db,
		settingsObj.RedisReader.PoolSize,
		settingsObj.RedisReader.Password,
	)

	writeRedisClient := redisutils.InitWriterRedisClient(
		settingsObj.Redis.Host,
		settingsObj.Redis.Port,
		settingsObj.Redis.Db,
		settingsObj.Redis.PoolSize,
		settingsObj.Redis.Password,
	)

	reporter := reporting.InitIssueReporter(settingsObj)

	caching.NewRedisCache(readRedisClient, writeRedisClient)
	smartcontract.InitContractAPI()
	taskmgr.NewRabbitmqTaskMgr()
	w3storage.InitW3S()
	caching.InitDiskCache()

	service.InitPayloadCommitService(reporter)

	mqWorker := worker.NewWorker()

	// health check is non-blocking health check http listener
	health.HealthCheck()

	defer func() {
		mqWorker.ShutdownWorker()

		err := readRedisClient.Close()
		if err != nil {
			log.WithError(err).Error("error while closing redis client")
		}

		err = writeRedisClient.Close()
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

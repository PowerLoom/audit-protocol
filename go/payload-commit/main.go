package main

import (
	log "github.com/sirupsen/logrus"

	"audit-protocol/caching"
	"audit-protocol/goutils/ethclient"
	"audit-protocol/goutils/health"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/reporting"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/smartcontract"
	"audit-protocol/goutils/smartcontract/transactions"
	taskmgr "audit-protocol/goutils/taskmgr/rabbitmq"
	w3storage "audit-protocol/goutils/w3s"
	"audit-protocol/payload-commit/service"
	"audit-protocol/payload-commit/signer"
	"audit-protocol/payload-commit/worker"
)

func main() {
	logger.InitLogger()

	settingsObj := settings.ParseSettings()

	ipfsService := ipfsutils.InitService(settingsObj)

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

	ethService, ethClient := ethclient.NewClient(settingsObj)
	reporter := reporting.InitIssueReporter(settingsObj)

	contractAPIService := smartcontract.InitContractAPI(settingsObj.Signer.Domain.VerifyingContract, ethClient)
	redisCache := caching.NewRedisCache(readRedisClient, writeRedisClient)

	rabbitmqTaskMgrConn, err := taskmgr.Dial(settingsObj)
	if err != nil {
		log.WithError(err).Fatal("failed to dial rabbitmq")
	}

	rabbitmqTaskMgr := taskmgr.NewRabbitmqTaskMgr(settingsObj, rabbitmqTaskMgrConn)
	web3StorageService := w3storage.InitW3S(settingsObj)
	diskCacheService := caching.InitDiskCache()
	signerService := signer.InitService(settingsObj)
	txManagerService := transactions.NewTxManager(settingsObj, ethClient, contractAPIService)

	payloadCommitService := service.InitPayloadCommitService(
		settingsObj,
		redisCache,
		ethService,
		reporter,
		ipfsService,
		diskCacheService,
		web3StorageService,
		contractAPIService,
		signerService,
		txManagerService,
	)

	mqWorker := worker.NewWorker(settingsObj, payloadCommitService, rabbitmqTaskMgr)

	// health check is non-blocking health check http listener
	health.HealthCheck()

	defer func() {
		err := mqWorker.ShutdownWorker()
		if err != nil {
			log.WithError(err).Error("error while shutting down worker")
		}

		err = readRedisClient.Close()
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

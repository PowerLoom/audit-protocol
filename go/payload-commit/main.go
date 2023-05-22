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
	taskmgr "audit-protocol/goutils/taskmgr/rabbitmq"
	w3storage "audit-protocol/goutils/w3s"
	"audit-protocol/payload-commit/service"
	"audit-protocol/payload-commit/signer"
	"audit-protocol/payload-commit/worker"
)

func main() {
	logger.InitLogger()

	settingsObj := settings.ParseSettings()

	ipfsService := ipfsutils.InitService(
		settingsObj.IpfsConfig.URL,
		settingsObj,
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

	ethService, ethClient := ethclient.NewClient(settingsObj.AnchorChainRPCURL)
	reporter := reporting.InitIssueReporter(settingsObj)

	redisCache := caching.NewRedisCache(redisClient)
	contractAPIService := smartcontract.InitContractAPI(settingsObj.Signer.Domain.VerifyingContract, ethClient)

	rabbitmqTaskMgrConn, err := taskmgr.Dial(settingsObj)
	if err != nil {
		log.WithError(err).Fatal("failed to dial rabbitmq")
	}

	rabbitmqTaskMgr := taskmgr.NewRabbitmqTaskMgr(settingsObj, rabbitmqTaskMgrConn)
	web3StorageService := w3storage.InitW3S(settingsObj)
	diskCacheService := caching.InitDiskCache()
	signerService := signer.InitService(settingsObj)

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
	)

	mqWorker := worker.NewWorker(settingsObj, payloadCommitService, rabbitmqTaskMgr)

	// health check is non-blocking health check http listener
	health.HealthCheck()

	defer func() {
		err := mqWorker.ShutdownWorker()
		if err != nil {
			log.WithError(err).Error("error while shutting down worker")
		}

		err = redisClient.Close()
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

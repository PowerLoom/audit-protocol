package main

import (
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/settings"
	pruning "audit-protocol/pruning/service"
)

func main() {
	logger.InitLogger()

	settingsObj := settings.ParseSettings()

	ipfsClient := ipfsutils.InitClient(
		settingsObj.IpfsConfig.URL,
		settingsObj.IpfsConfig.IPFSRateLimiter,
		settingsObj.IpfsConfig.Timeout,
	)

	cronRunner := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger),
	))

	// Run every 7days
	cronId, err := cronRunner.AddFunc(settingsObj.Pruning.CronFrequency, func() {
		pruning.Prune(settingsObj, ipfsClient)
	})
	if err != nil {
		log.WithError(err).Fatal("failed to add pruning cron job")
	}

	log.WithField("cronId", cronId).Info("added pruning cron job")

	cronRunner.Start()

	// block forever
	select {}
}

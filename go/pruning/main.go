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

	ipfsService := ipfsutils.InitService(settingsObj)

	cronRunner := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger),
	))

	cronId, err := cronRunner.AddFunc(settingsObj.Pruning.CronFrequency, func() {
		pruning.Prune(settingsObj, ipfsService)
	})
	if err != nil {
		log.WithError(err).Fatal("failed to add pruning cron job")
	}

	log.WithField("cronId", cronId).Info("added pruning cron job")

	cronRunner.Start()

	// block forever
	select {}
}

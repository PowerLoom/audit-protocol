package smartcontract

import (
    log "github.com/sirupsen/logrus"
    "github.com/swagftw/gi"

    "audit-protocol/goutils/settings"
)

// Keeping this util straight forward and simple for now

func InitEthClient() {
    settingsObj, err := gi.Invoke[*settings.SettingsObj]()

    if err != nil {
        log.Fatal("failed to invoke settings object")
    }
}

func getEthClient() {

}

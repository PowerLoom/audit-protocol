package smartcontract

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/settings"
	contractApi "audit-protocol/goutils/smartcontract/api"
)

// Keeping this util straight forward and simple for now

func InitContractAPI() *contractApi.ContractApi {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()

	if err != nil {
		log.Fatal("failed to invoke settings object")
	}

	client, err := ethclient.Dial(settingsObj.AnchorChainRPCURL)
	if err != nil {
		log.WithError(err).Fatal("failed to init eth client")
	}

	err = gi.Inject(client)
	if err != nil {
		log.WithError(err).Fatal("failed to inject eth client")
	}

	apiConn, err := contractApi.NewContractApi(common.HexToAddress(settingsObj.Signer.Domain.VerifyingContract), client)
	if err != nil {
		log.WithError(err).Fatal("failed to init api connection")
	}

	err = gi.Inject(apiConn)
	if err != nil {
		log.WithError(err).Fatal("failed to inject eth client")
	}

	return apiConn
}

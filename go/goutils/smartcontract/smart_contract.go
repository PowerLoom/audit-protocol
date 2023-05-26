package smartcontract

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
	contractApi "audit-protocol/goutils/smartcontract/api"
)

// Keeping this util straight forward and simple for now

func InitContractAPI() *contractApi.ContractApi {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()

	if err != nil {
		log.Fatal("failed to invoke settings object")
	}

	httpClient := httpclient.GetDefaultHTTPClient()

	rpClient, err := rpc.DialOptions(context.Background(), settingsObj.AnchorChainRPCURL, rpc.WithHTTPClient(httpClient.HTTPClient))
	if err != nil {
		log.WithError(err).Fatal("failed to init rpc client")
	}

	ethClient := ethclient.NewClient(rpClient)

	err = gi.Inject(ethClient)
	if err != nil {
		log.WithError(err).Fatal("failed to inject eth client")
	}

	apiConn, err := contractApi.NewContractApi(common.HexToAddress(settingsObj.Signer.Domain.VerifyingContract), ethClient)
	if err != nil {
		log.WithError(err).Fatal("failed to init api connection")
	}

	err = gi.Inject(apiConn)
	if err != nil {
		log.WithError(err).Fatal("failed to inject eth client")
	}

	return apiConn
}

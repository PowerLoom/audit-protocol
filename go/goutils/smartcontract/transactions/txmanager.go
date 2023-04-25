package transactions

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/settings"
)

type NonceManager struct {
	Mu    sync.Mutex
	Nonce int
}

func NewNonceManager() *NonceManager {
	return &NonceManager{
		Mu:    sync.Mutex{},
		Nonce: int(getNonce()),
	}
}

func getNonce() uint64 {
	ethClient, _ := gi.Invoke[*ethclient.Client]()
	settingsObj, _ := gi.Invoke[*settings.SettingsObj]()

	nonce, err := ethClient.PendingNonceAt(context.Background(), common.HexToAddress(settingsObj.Signer.AccountAddress))
	if err != nil {
		log.WithError(err).Fatal("failed to get pending transaction count")
	}

	return nonce
}

func CreateAndSendRawTransaction(client *ethclient.Client, privKey *ecdsa.PrivateKey, settingsObj *settings.SettingsObj, nonce uint64, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.WithError(err).Error("failed to marshal data payload to be sent to transaction")

		return err
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.WithError(err).Error("failed to get chainID")
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	toAddress := common.HexToAddress(settingsObj.Signer.Domain.VerifyingContract)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      uint64(2000000),
		To:       &toAddress,
		Value:    big.NewInt(0),
		Data:     dataBytes,
	})

	// sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privKey)
	if err != nil {
		log.WithError(err).Error("failed to sign transaction")

		return err
	}

	// send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.WithError(err).Error("failed to send transaction")

		return err
	}

	return nil
}

package transactions

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/settings"
	contractApi "audit-protocol/goutils/smartcontract/api"
)

type TxManager struct {
	Mu          sync.Mutex
	Nonce       uint64
	ChainID     *big.Int
	settingsObj *settings.SettingsObj
	ethClient   *ethclient.Client
	gasPrice    *big.Int
}

func NewNonceManager() *TxManager {
	ethClient, err := gi.Invoke[*ethclient.Client]()
	if err != nil {
		log.Fatal("failed to invoke eth client")
	}

	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.Fatal("failed to invoke settings object")
	}

	chainId, err := ethClient.ChainID(context.Background())
	if err != nil {
		log.WithError(err).Fatal("failed to get chain id")
	}

	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		gasPrice = big.NewInt(10000000)
	}

	txMgr := &TxManager{
		Mu:          sync.Mutex{},
		ChainID:     chainId,
		settingsObj: settingsObj,
		ethClient:   ethClient,
		gasPrice:    gasPrice,
	}

	txMgr.Nonce = txMgr.getNonce(ethClient)

	return txMgr
}

// getNonce gets the nonce for the account
func (t *TxManager) getNonce(ethClient *ethclient.Client) uint64 {
	nonce, err := ethClient.PendingNonceAt(context.Background(), common.HexToAddress(t.settingsObj.Signer.AccountAddress))
	if err != nil {
		log.WithError(err).Fatal("failed to get pending transaction count")
	}

	return nonce
}

// SubmitSnapshot submits a snapshot to the smart contract
func (t *TxManager) SubmitSnapshot(api *contractApi.ContractApi, privKey *ecdsa.PrivateKey, signerData *apitypes.TypedData, msg *datamodel.SnapshotRelayerPayload, signature []byte) error {
	t.Mu.Lock()

	deadline := signerData.Message["deadline"].(*math.HexOrDecimal256)

	var submitSnapshotErr error
	var signedTx *types.Transaction

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func(nonce uint64) {
		defer wg.Done()

		signedTx, submitSnapshotErr = api.SubmitSnapshot(
			&bind.TransactOpts{
				Nonce:    big.NewInt(int64(nonce)),
				Value:    big.NewInt(0),
				GasPrice: t.gasPrice,
				GasLimit: 2000000,
				From:     common.HexToAddress(t.settingsObj.Signer.AccountAddress),
				Signer: func(address common.Address, transaction *types.Transaction) (*types.Transaction, error) {
					signedTx, err := types.SignTx(transaction, types.NewEIP155Signer(t.ChainID), privKey)
					if err != nil {
						log.WithError(err).Error("failed to sign transaction for snapshot")

						return nil, err
					}

					return signedTx, nil
				},
			},
			msg.SnapshotCID,
			big.NewInt(int64(msg.EpochID)),
			msg.ProjectID,
			contractApi.PowerloomProtocolStateRequest{
				Deadline:    (*big.Int)(deadline),
				SnapshotCid: msg.SnapshotCID,
				EpochId:     big.NewInt(int64(msg.EpochID)),
				ProjectId:   msg.ProjectID,
			},
			signature,
		)

		if submitSnapshotErr == nil && signedTx != nil {
			log.WithField("txHash", signedTx.Hash().Hex()).Info("snapshot submitted successfully")
		}

	}(t.Nonce)

	t.Nonce++
	t.Mu.Unlock()

	wg.Wait()

	if submitSnapshotErr != nil {
		log.WithError(submitSnapshotErr).Error("failed to submit snapshot")

		return submitSnapshotErr
	}

	return nil
}

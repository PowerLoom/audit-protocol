package transactions

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ethclient"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/smartcontract"
)

type Service interface {
	SubmitSnapshotToContract(privKey *ecdsa.PrivateKey, signerData *apitypes.TypedData, msg *datamodel.SnapshotRelayerPayload, signature []byte) error
	SendSignatureToRelayer(payload *datamodel.SnapshotRelayerPayload) error
}

type TxManager struct {
	Mu                 sync.Mutex
	Nonce              uint64
	ChainID            *big.Int
	settingsObj        *settings.SettingsObj
	ethService         ethclient.Service
	contractAPIService smartcontract.Service
	gasPrice           *big.Int
}

func NewTxManager(settingsObj *settings.SettingsObj, ethService ethclient.Service, contractApiService smartcontract.Service) Service {
	chainId, err := ethService.ChainID(context.Background())
	if err != nil {
		log.WithError(err).Fatal("failed to get chain id")
	}

	gasPrice, err := ethService.SuggestGasPrice(context.Background())
	if err != nil {
		gasPrice = big.NewInt(10000000)
	}

	txMgr := &TxManager{
		Mu:                 sync.Mutex{},
		ChainID:            chainId,
		settingsObj:        settingsObj,
		ethService:         ethService,
		contractAPIService: contractApiService,
		gasPrice:           gasPrice,
	}

	txMgr.Nonce, err = ethService.PendingNonceAt(context.Background(), common.HexToAddress(settingsObj.Signer.AccountAddress))
	if err != nil {
		log.WithError(err).Fatal("failed to get nonce")
	}

	return txMgr
}

// SubmitSnapshotToContract submits a snapshot to the smart contract
func (t *TxManager) SubmitSnapshotToContract(privKey *ecdsa.PrivateKey, signerData *apitypes.TypedData, msg *datamodel.SnapshotRelayerPayload, signature []byte) error {
	t.Mu.Lock()

	deadline, ok := signerData.Message["deadline"].(*math.HexOrDecimal256)
	if !ok {
		return errors.New("failed to get deadline from EIP712 message")
	}

	var submitSnapshotErr error
	var signedTx *types.Transaction

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(nonce uint64) {
		defer wg.Done()
		signedTx, submitSnapshotErr = t.contractAPIService.SubmitSnapshotToContract(nonce, privKey, t.gasPrice, t.ChainID, t.settingsObj.Signer.AccountAddress, deadline, msg, signature)

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

func (t *TxManager) SendSignatureToRelayer(payload *datamodel.SnapshotRelayerPayload) error {
	rb := &datamodel.RelayerRequest{
		ProjectID:   payload.ProjectID,
		SnapshotCID: payload.SnapshotCID,
		EpochID:     payload.EpochID,
		Request: &datamodel.Request{
			Deadline: (*big.Int)(payload.Request["deadline"].(*math.HexOrDecimal256)).Uint64(),
		},
		Signature: "0x" + payload.Signature,
	}

	httpClient := httpclient.GetDefaultHTTPClient(t.settingsObj.Relayer.Timeout, t.settingsObj)

	// url = "host+port" ; endpoint = "/endpoint"
	url := *t.settingsObj.Relayer.Host + *t.settingsObj.Relayer.Endpoint

	payloadBytes, err := json.Marshal(rb)
	if err != nil {
		log.WithError(err).Error("failed to marshal payload")

		return err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.WithError(err).Error("failed to create request to relayer")

		return err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to send request to relayer")

		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.WithField("status", res.StatusCode).Error("failed to send request to relayer")

		return errors.New("failed to send request to relayer, status not ok")
	}

	log.Info("successfully sent signature to relayer")

	return nil
}

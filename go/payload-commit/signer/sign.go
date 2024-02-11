package signer

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	types "github.com/ethereum/go-ethereum/signer/core/apitypes"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/settings"
)

func SignMessage(privKey *ecdsa.PrivateKey, signerData *types.TypedData) ([]byte, error) {
	data, _, err := types.TypedDataAndHash(*signerData)
	if err != nil {
		log.WithError(err).Error("failed to encode for signing")

		return nil, err
	}

	sig, err := crypto.Sign(data, privKey)
	if err != nil {
		log.WithError(err).Error("failed to sign message")

		return nil, err
	}

	sig[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	finalSig := sig

	log.Info("final signature (hex):", hex.EncodeToString(finalSig))

	return finalSig, nil
}

func GetSignerData(client *ethclient.Client, snapshotCid, projectId string, epochId int64) (*types.TypedData, error) {
	log.Debug("getting signer data")
	var block uint64
	var err error

	settingsObj, _ := gi.Invoke[*settings.SettingsObj]()

	err = backoff.Retry(func() error {
		block, err = client.BlockNumber(context.Background())
		if err != nil {
			log.WithError(err).Error("failed to get block number")

			return err
		}

		log.Info("block number fetched: ", block)

		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3))
	if err != nil {
		log.WithError(err).Error("failed to get block number after max retries")

		return nil, err
	}

	signerData := &types.TypedData{
		PrimaryType: "Request",
		Types: types.Types{
			"Request": []types.Type{
				{Name: "slotId", Type: "uint256"}
				{Name: "deadline", Type: "uint256"},
				{Name: "snapshotCid", Type: "string"},
				{Name: "epochId", Type: "uint256"},
				{Name: "projectId", Type: "string"},
			},
			"EIP712Domain": []types.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		Domain: types.TypedDataDomain{
			Name:              settingsObj.Signer.Domain.Name,
			Version:           settingsObj.Signer.Domain.Version,
			ChainId:           (*math.HexOrDecimal256)(math.MustParseBig256(settingsObj.Signer.Domain.ChainId)),
			VerifyingContract: settingsObj.Signer.Domain.VerifyingContract,
		},
		Message: types.TypedDataMessage{
			"slotId":      (*math.HexOrDecimal256)(big.NewInt(0)),
			"deadline":    (*math.HexOrDecimal256)(big.NewInt(int64(block) + int64(settingsObj.Signer.DeadlineBuffer))),
			"snapshotCid": snapshotCid,
			"epochId":     (*math.HexOrDecimal256)(big.NewInt(epochId)),
			"projectId":   projectId,
		},
	}

	go func() {
		data, _ := json.Marshal(signerData)
		log.Info("signer data: ", string(data))
	}()

	return signerData, nil
}

func GetPrivateKey(privateKey string) (*ecdsa.PrivateKey, error) {
	pkBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		log.WithError(err).Error("failed to decode private key")

		return nil, err
	}

	pk, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		log.WithError(err).Error("failed to convert private key to ECDSA")

		return nil, err
	}

	return pk, nil
}

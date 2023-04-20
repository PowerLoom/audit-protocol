package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	types "github.com/ethereum/go-ethereum/signer/core/apitypes"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/settings"
)

func signMessage(privKey *ecdsa.PrivateKey, signerData *types.TypedData) ([]byte, error) {
	message, err := json.Marshal(signerData)
	if err != nil {
		log.WithError(err).Error("failed to marshal signer data")

		return nil, err
	}

	hashedMessage := crypto.Keccak256Hash(message)

	sig, err := crypto.Sign(hashedMessage.Bytes(), privKey)
	if err != nil {
		log.Fatal(err)
	}

	v, r, s := sig[64]+27, sig[:32], sig[32:64]
	finalSig := append(append(r, s...), v)

	log.Info("final signature:", finalSig)
	log.Info("final signature (hex):", hex.EncodeToString(finalSig))

	return finalSig, nil
}

func VerifySignature(signature []byte, signerData *types.TypedData) bool {
	// check length of signature
	if len(signature) != 65 {
		log.Error("invalid signature length")

		return false
	}

	// check if signature is valid
	if signature[64] != 27 && signature[64] != 28 {
		log.Errorf("invalid recovery id: %d", signature[64])

		return false
	}

	typedDataHash, err := signerData.HashStruct(signerData.PrimaryType, signerData.Message)
	if err != nil {
		log.Fatal(err)
	}

	domainSeparator, err := signerData.HashStruct("EIP712Domain", signerData.Domain.Map())
	if err != nil {
		log.Fatal(err)
	}

	data := fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash))
	messageHash := crypto.Keccak256Hash([]byte(data))

	signature[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	pubKeyRaw, err := crypto.Ecrecover(messageHash.Bytes(), signature)
	if err != nil {
		log.WithError(err).Error("failed to recover public key")

		return false
	}

	if !crypto.VerifySignature(pubKeyRaw, messageHash[:], signature[:64]) {
		log.Error("verification failed, addresses do not match")

		return false
	}

	return true
}

func GetSignerData() (*types.TypedData, error) {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.Fatal("failed to invoke settings object")
	}

	client, err := ethclient.Dial(settingsObj.Signer.RPCUrl)
	if err != nil {
		log.WithError(err).Error("failed to connect to ethereum node")

		return nil, err
	}

	var block uint64

	err = backoff.Retry(func() error {
		block, err = client.BlockNumber(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3))
	if err != nil {
		log.WithError(err).Error("failed to get block number")

		return nil, err
	}

	signerData := &types.TypedData{
		Types: types.Types{
			"Request": []types.Type{
				{Name: "deadline", Type: "string"},
			},
			"EIP712Domain": []types.Type{
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "version", Type: "string"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "Request",
		Domain: types.TypedDataDomain{
			Name:              settingsObj.Signer.Domain.Name,
			Version:           settingsObj.Signer.Domain.Version,
			ChainId:           (*math.HexOrDecimal256)(math.MustParseBig256(settingsObj.Signer.Domain.ChainID)),
			VerifyingContract: settingsObj.Signer.Domain.VerifyingContract,
		},
		Message: types.TypedDataMessage{
			"deadline": strconv.Itoa(int(block + settingsObj.Signer.DeadlineBuffer)),
		},
	}

	return signerData, nil
}

func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Error("PRIVATE_KEY env variable is not set")

		return nil, fmt.Errorf("PRIVATE_KEY env variable is not set")
	}

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

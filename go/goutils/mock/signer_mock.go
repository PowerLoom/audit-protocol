package mock

import (
	"crypto/ecdsa"

	types "github.com/ethereum/go-ethereum/signer/core/apitypes"

	"audit-protocol/goutils/ethclient"
)

type SignerMock struct {
	SignMessageMock   func(privKey *ecdsa.PrivateKey, signerData *types.TypedData) ([]byte, error)
	GetSignerDataMock func(ethService ethclient.Service, snapshotCid, projectId string, epochId int64) (*types.TypedData, error)
	GetPrivateKeyMock func(privateKey string) (*ecdsa.PrivateKey, error)
}

func (s *SignerMock) SignMessage(privKey *ecdsa.PrivateKey, signerData *types.TypedData) ([]byte, error) {
	return s.SignMessageMock(privKey, signerData)
}

func (s *SignerMock) GetSignerData(ethService ethclient.Service, snapshotCid, projectId string, epochId int64) (*types.TypedData, error) {
	return s.GetSignerDataMock(ethService, snapshotCid, projectId, epochId)
}

func (s *SignerMock) GetPrivateKey(privateKey string) (*ecdsa.PrivateKey, error) {
	return s.GetPrivateKeyMock(privateKey)
}

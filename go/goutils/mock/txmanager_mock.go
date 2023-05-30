package mock

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"audit-protocol/goutils/datamodel"
)

type TxManagerMock struct {
	SubmitSnapshotToContractMock func(privKey *ecdsa.PrivateKey, signerData *apitypes.TypedData, msg *datamodel.SnapshotRelayerPayload, signature []byte) error
	SendSignatureToRelayerMock   func(payload *datamodel.SnapshotRelayerPayload) error
}

func (m TxManagerMock) SubmitSnapshotToContract(privKey *ecdsa.PrivateKey, signerData *apitypes.TypedData, msg *datamodel.SnapshotRelayerPayload, signature []byte) error {
	return m.SubmitSnapshotToContractMock(privKey, signerData, msg, signature)
}

func (m TxManagerMock) SendSignatureToRelayer(payload *datamodel.SnapshotRelayerPayload) error {
	return m.SendSignatureToRelayerMock(payload)
}

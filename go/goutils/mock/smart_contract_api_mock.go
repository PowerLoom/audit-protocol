package mock

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"

	"audit-protocol/goutils/datamodel"
)

type SmartContractAPIMock struct {
	SubmitSnapshotToContractMock func(nonce uint64, privKey *ecdsa.PrivateKey, gasPrice, chainId *big.Int, accountAddress string, deadline *math.HexOrDecimal256, msg *datamodel.SnapshotRelayerPayload, signature []byte) (*types.Transaction, error)
	GetProjectsMock              func() ([]string, error)
}

func (m SmartContractAPIMock) SubmitSnapshotToContract(nonce uint64, privKey *ecdsa.PrivateKey, gasPrice, chainId *big.Int, accountAddress string, deadline *math.HexOrDecimal256, msg *datamodel.SnapshotRelayerPayload, signature []byte) (*types.Transaction, error) {
	return m.SubmitSnapshotToContractMock(nonce, privKey, gasPrice, chainId, accountAddress, deadline, msg, signature)
}

func (m SmartContractAPIMock) GetProjects() ([]string, error) {
	return m.GetProjectsMock()
}

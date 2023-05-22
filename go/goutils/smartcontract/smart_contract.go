package smartcontract

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/datamodel"
	contractApi "audit-protocol/goutils/smartcontract/api"
)

type Service interface {
	SubmitSnapshotToContract(nonce uint64, privKey *ecdsa.PrivateKey, gasPrice, chainId *big.Int, accountAddress string, deadline *math.HexOrDecimal256, msg *datamodel.SnapshotRelayerPayload, signature []byte) (*types.Transaction, error)
	GetProjects() ([]string, error)
}

type ContractApi struct {
	conn *contractApi.ContractApi
}

func (c ContractApi) SubmitSnapshotToContract(nonce uint64, privKey *ecdsa.PrivateKey, gasPrice, chainId *big.Int, accountAddress string, deadline *math.HexOrDecimal256, msg *datamodel.SnapshotRelayerPayload, signature []byte) (*types.Transaction, error) {
	signedTx, err := c.conn.SubmitSnapshot(
		&bind.TransactOpts{
			Nonce:    big.NewInt(int64(nonce)),
			Value:    big.NewInt(0),
			GasPrice: gasPrice,
			GasLimit: 2000000,
			From:     common.HexToAddress(accountAddress),
			Signer: func(address common.Address, transaction *types.Transaction) (*types.Transaction, error) {
				signedTx, err := types.SignTx(transaction, types.NewEIP155Signer(chainId), privKey)
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

	return signedTx, err
}

func (c ContractApi) GetProjects() ([]string, error) {
	return c.conn.GetProjects(&bind.CallOpts{})
}

func InitContractAPI(contractAddress string, client *ethclient.Client) Service {
	apiConn, err := contractApi.NewContractApi(common.HexToAddress(contractAddress), client)
	if err != nil {
		log.WithError(err).Fatal("failed to init api connection")
	}

	return &ContractApi{
		conn: apiConn,
	}
}

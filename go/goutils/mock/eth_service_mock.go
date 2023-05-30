package mock

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type EthServiceMock struct {
	ChainIDMock         func(ctx context.Context) (*big.Int, error)
	PendingNonceAtMock  func(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPriceMock func(ctx context.Context) (*big.Int, error)
	BlockNumberMock     func(ctx context.Context) (uint64, error)
}

func (m EthServiceMock) ChainID(ctx context.Context) (*big.Int, error) {
	return m.ChainIDMock(ctx)
}

func (m EthServiceMock) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return m.PendingNonceAtMock(ctx, account)
}

func (m EthServiceMock) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return m.SuggestGasPriceMock(ctx)
}

func (m EthServiceMock) BlockNumber(ctx context.Context) (uint64, error) {
	return m.BlockNumberMock(ctx)
}

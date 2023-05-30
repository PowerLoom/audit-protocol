package ethclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
)

type Service interface {
	ChainID(ctx context.Context) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
}

type Client struct {
	client *ethclient.Client
}

func NewClient(settingsObj *settings.SettingsObj) (Service, *ethclient.Client) {
	httpClient := httpclient.GetDefaultHTTPClient(settingsObj.HttpClient.ConnectionTimeout, settingsObj)

	rpClient, err := rpc.DialOptions(context.Background(), settingsObj.AnchorChainRPCURL, rpc.WithHTTPClient(httpClient.HTTPClient))
	if err != nil {
		log.WithError(err).Fatal("failed to init rpc client")
	}

	ethClient := ethclient.NewClient(rpClient)

	return &Client{client: ethClient}, ethClient
}

func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.client.BlockNumber(ctx)
	if err != nil {
		log.WithError(err).Error("failed to get block number")
	}

	return blockNumber, err
}

func (c *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		log.WithError(err).Error("failed to get gas price")
	}

	return gasPrice, err
}

func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	chainId, err := c.client.ChainID(ctx)
	if err != nil {
		log.WithError(err).Error("failed to get chain id")
	}

	return chainId, err
}

func (c *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nonce, err := c.client.PendingNonceAt(ctx, account)
	if err != nil {
		log.WithError(err).Error("failed to get nonce")
	}

	return nonce, err
}

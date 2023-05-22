package service

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"audit-protocol/caching"
	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ethclient"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/reporting"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/smartcontract"
	"audit-protocol/goutils/taskmgr"
	w3storage "audit-protocol/goutils/w3s"
	"audit-protocol/payload-commit/signer"
)

func initSettings() *settings.SettingsObj {
	return &settings.SettingsObj{
		InstanceId:        "",
		PoolerNamespace:   "",
		AnchorChainRPCURL: "",
		LocalCachePath:    "",
		Concurrency:       0,
		WorkerConcurrency: 0,
		HttpClient:        nil,
		Rabbitmq:          nil,
		IpfsConfig:        nil,
		Redis:             nil,
		RedisReader:       nil,
		Web3Storage:       nil,
		Signer: &settings.Signer{
			Domain: struct {
				Name              string `json:"name"`
				Version           string `json:"version"`
				ChainId           string `json:"chainId"`
				VerifyingContract string `json:"verifyingContract"`
			}{
				Name:              "",
				Version:           "",
				ChainId:           "",
				VerifyingContract: "",
			},
			AccountAddress: "",
			PrivateKey:     "",
			DeadlineBuffer: 0,
		},
		Relayer:   nil,
		Pruning:   nil,
		Reporting: nil,
	}
}

type MockRedisCache struct{}

func (m MockRedisCache) GetSnapshotAtEpochID(ctx context.Context, projectID string, epochId int) (*datamodel.UnfinalizedSnapshot, error) {
	return &datamodel.UnfinalizedSnapshot{
		SnapshotCID: "snapshotCID",
		Snapshot: map[string]interface{}{
			"snapshot": "snapshot",
		},
		TTL: 123,
	}, nil
}

var _ caching.DbCache = &MockRedisCache{}

func (m MockRedisCache) GetStoredProjects(ctx context.Context) ([]string, error) {
	return []string{"project1", "project2"}, nil
}

func (m MockRedisCache) CheckIfProjectExists(ctx context.Context, projectID string) (bool, error) {
	return true, nil
}

func (m MockRedisCache) StoreProjects(background context.Context, projects []string) error {
	return nil
}

func (m MockRedisCache) AddUnfinalizedSnapshotCID(ctx context.Context, msg *datamodel.PayloadCommitMessage, ttl int64) error {
	return nil
}

func (m MockRedisCache) AddSnapshotterStatusReport(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport) error {
	return nil
}

type MockEthClient struct{}

func (m MockEthClient) ChainID(ctx context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
}

func (m MockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return 1, nil
}

func (m MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
}

func (m MockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	return 1, nil
}

var _ ethclient.Service = &MockEthClient{}

type MockReporter struct{}

func (m MockReporter) Report(issueType reporting.IssueType, projectID string, epochID string, extra map[string]interface{}) {
	return
}

var _ reporting.Service = &MockReporter{}

type MockIPFSService struct{}

func (m MockIPFSService) UploadSnapshotToIPFS(snapshot *datamodel.PayloadCommitMessage) error {
	return nil
}

func (m MockIPFSService) GetSnapshotFromIPFS(snapshotCID string, outputPath string) error {
	return nil
}

func (m MockIPFSService) Unpin(cid string) error {
	return nil
}

var _ ipfsutils.Service = &MockIPFSService{}

type MockDiskCache struct{}

func (m MockDiskCache) Read(filepath string) ([]byte, error) {
	return []byte("hello world!"), nil

}

func (m MockDiskCache) Write(filepath string, data []byte) error {
	return nil
}

var _ caching.DiskCache = &MockDiskCache{}

type MockWeb3Storage struct{}

func (m MockWeb3Storage) UploadToW3s(msg interface{}) (string, error) {
	return "w3sCID", nil
}

var _ w3storage.Service = &MockWeb3Storage{}

type MockContractAPIService struct{}

func (m MockContractAPIService) SubmitSnapshotToContract(nonce uint64, privKey *ecdsa.PrivateKey, gasPrice, chainId *big.Int, accountAddress string, deadline *math.HexOrDecimal256, msg *datamodel.SnapshotRelayerPayload, signature []byte) (*types.Transaction, error) {
	to := common.HexToAddress("0x6FC64a0356Ebb41F88cd31e69A5C16e35DFFaDd1")
	return types.NewTx(&types.LegacyTx{
		Nonce:    1,
		GasPrice: big.NewInt(1),
		Gas:      1,
		To:       &to,
		Value:    big.NewInt(1),
		Data:     nil,
		V:        nil,
		R:        nil,
		S:        nil,
	}), nil
}

func (m MockContractAPIService) GetProjects() ([]string, error) {
	return []string{"project1", "project2"}, nil
}

var _ smartcontract.Service = &MockContractAPIService{}

type MockSignerService struct{}

func (m MockSignerService) SignMessage(privKey *ecdsa.PrivateKey, signerData *apitypes.TypedData) ([]byte, error) {
	return []byte("signature"), nil
}

func (m MockSignerService) GetSignerData(ethService ethclient.Service, snapshotCid, projectId string, epochId int64) (*apitypes.TypedData, error) {
	return &apitypes.TypedData{}, nil
}

func (m MockSignerService) GetPrivateKey(privateKey string) (*ecdsa.PrivateKey, error) {
	return nil, nil
}

var _ signer.Service = &MockSignerService{}

func initPayloadCommitService() Service {
	settingsObj := initSettings()

	mockRedisCache := new(MockRedisCache)
	mockEthClientService := new(MockEthClient)
	mockReportingService := new(MockReporter)
	mockIPfsService := new(MockIPFSService)
	mockDiskCache := new(MockDiskCache)
	mockWeb3Storage := new(MockWeb3Storage)
	mockContractAPIService := new(MockContractAPIService)
	mockSignerService := new(MockSignerService)

	return InitPayloadCommitService(settingsObj, mockRedisCache, mockEthClientService, mockReportingService, mockIPfsService, mockDiskCache, mockWeb3Storage, mockContractAPIService, mockSignerService)
}

func TestPayloadCommitService_Run(t *testing.T) {
	pcService := initPayloadCommitService()

	payloadCommitMsg := &datamodel.PayloadCommitMessage{
		Message:       map[string]interface{}{"hello": "world"},
		Web3Storage:   false,
		SourceChainID: 100,
		ProjectID:     "project1",
		EpochID:       1,
		SnapshotCID:   "snapshotCID",
	}

	payloadCommitMsgJsonData, _ := json.Marshal(payloadCommitMsg)

	payloadCommitFinalizedMsg := &datamodel.PayloadCommitFinalizedMessage{
		Message: &datamodel.PowerloomSnapshotFinalizedMessage{
			EpochID:     1,
			ProjectID:   "project1",
			SnapshotCID: "snapshotCID",
			Timestamp:   int(time.Now().Unix()),
		},
		Web3Storage:   false,
		SourceChainID: 100,
	}

	payloadCommitFinalizedMsgJsonData, _ := json.Marshal(payloadCommitFinalizedMsg)

	type args struct {
		msgBody []byte
		topic   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should return error when msgBody is nil",
			args: args{
				msgBody: nil,
				topic:   taskmgr.DataSuffix,
			},
			wantErr: true,
		},
		{
			name: "should not return error when topic is not matched",
			args: args{
				msgBody: []byte("hello world!"),
				topic:   "random-topic",
			},
			wantErr: false,
		},
		{
			name: "should return error when msgBody is not valid json",
			args: args{
				msgBody: []byte("hello world!"),
				topic:   taskmgr.DataSuffix,
			},
			wantErr: true,
		},
		{
			name: "should return error when msgBody is not valid json with .Data topic",
			args: args{
				msgBody: []byte("{\"hello\":\"world\"}"),
				topic:   taskmgr.DataSuffix,
			},
			wantErr: true,
		},
		{
			name: "should return error when msgBody is not valid json with .Finalized topic",
			args: args{
				msgBody: []byte("{\"hello\":\"world\"}"),
				topic:   taskmgr.FinalizedSuffix,
			},
			wantErr: true,
		},
		{
			name: "should not return error when msgBody is valid json with .Data topic",
			args: args{
				msgBody: payloadCommitMsgJsonData,
				topic:   taskmgr.DataSuffix,
			},
			wantErr: false,
		},
		{
			name: "should not return error when msgBody is valid json with .Finalized topic",
			args: args{
				msgBody: payloadCommitFinalizedMsgJsonData,
				topic:   taskmgr.FinalizedSuffix,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := pcService.Run(tt.args.msgBody, tt.args.topic); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ethclient"
	"audit-protocol/goutils/mock"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/taskmgr"
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

func getPrivateKey() *ecdsa.PrivateKey {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	return privateKey
}

func TestPayloadCommitService_Run(t *testing.T) {
	settingsObj := initSettings()

	mockRedisCache := new(mock.RedisMock)
	mockEthClientService := new(mock.EthServiceMock)
	mockReportingService := new(mock.ReportingServiceMock)
	mockIPfsService := new(mock.IPFSServiceMock)
	mockDiskCache := new(mock.DiskMock)
	mockWeb3Storage := new(mock.W3SMock)
	mockContractAPIService := new(mock.SmartContractAPIMock)
	mockSignerService := new(mock.SignerMock)
	mockTxManager := new(mock.TxManagerMock)

	mockSignerService.GetPrivateKeyMock = func(string) (*ecdsa.PrivateKey, error) {
		// generate private key

		return getPrivateKey(), nil
	}

	mockEthClientService.ChainIDMock = func(context.Context) (*big.Int, error) {
		return big.NewInt(100), nil
	}

	mockEthClientService.SuggestGasPriceMock = func(context.Context) (*big.Int, error) {
		return big.NewInt(100), nil
	}

	mockEthClientService.PendingNonceAtMock = func(context.Context, common.Address) (uint64, error) {
		return 1, nil
	}

	mockRedisCache.GetStoredProjectsMock = func(context.Context) ([]string, error) {
		return []string{"project1"}, nil
	}

	mockRedisCache.GetSnapshotAtEpochIDMock = func(context.Context, string, int) (*datamodel.UnfinalizedSnapshot, error) {
		return &datamodel.UnfinalizedSnapshot{
			SnapshotCID: "snapshotCid",
			Snapshot:    map[string]interface{}{},
			TTL:         time.Now().Unix(),
		}, nil
	}

	mockIPfsService.UploadSnapshotToIPFSMock = func(message *datamodel.PayloadCommitMessage) error {
		return nil
	}

	mockWeb3Storage.UploadToW3sMock = func(interface{}) (string, error) {
		return "snapshotCid", nil
	}

	mockRedisCache.AddUnfinalizedSnapshotCIDMock = func(context.Context, *datamodel.PayloadCommitMessage, int64) error {
		return nil
	}

	mockSignerService.GetSignerDataMock = func(ethclient.Service, string, string, int64) (*apitypes.TypedData, error) {
		return &apitypes.TypedData{}, nil
	}

	mockSignerService.SignMessageMock = func(*ecdsa.PrivateKey, *apitypes.TypedData) ([]byte, error) {
		return []byte{}, nil
	}

	mockTxManager.SubmitSnapshotToContractMock = func(*ecdsa.PrivateKey, *apitypes.TypedData, *datamodel.SnapshotRelayerPayload, []byte) error {
		return nil
	}

	mockTxManager.SendSignatureToRelayerMock = func(*datamodel.SnapshotRelayerPayload) error {
		return nil
	}

	mockReportingService.ReportMock = func(issueType datamodel.IssueType, projectID string, epochID string, extra map[string]interface{}) {
		return
	}

	mockIPfsService.GetSnapshotFromIPFSMock = func(string, string) error {
		return nil
	}

	mockIPfsService.UnpinMock = func(string) error {
		return nil
	}

	mockRedisCache.AddSnapshotterStatusReportMock = func(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport) error {
		return nil
	}

	pcService := InitPayloadCommitService(settingsObj, mockRedisCache, mockEthClientService, mockReportingService, mockIPfsService, mockDiskCache, mockWeb3Storage, mockContractAPIService, mockSignerService, mockTxManager)

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

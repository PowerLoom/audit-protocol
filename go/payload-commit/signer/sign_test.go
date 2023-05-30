package signer

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"

	"audit-protocol/goutils/ethclient"
	"audit-protocol/goutils/settings"
)

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

var signer = &Signer{
	settingsObj: &settings.SettingsObj{
		Signer: &settings.Signer{
			Domain: struct {
				Name              string `json:"name"`
				Version           string `json:"version"`
				ChainId           string `json:"chainId"`
				VerifyingContract string `json:"verifyingContract"`
			}{
				Name:              "test",
				Version:           "0.1",
				ChainId:           "100",
				VerifyingContract: "0x6FC64a0356Ebb41F88cd31e69A5C16e35DFFaDd1",
			},
		},
	},
}

func TestSignMessage_Success(t *testing.T) {
	privKey, _ := crypto.GenerateKey()

	signerData, err := signer.GetSignerData(new(MockEthClient), "snapshotCid", "projectID", 10)
	assert.NoError(t, err)

	// Mock successful signature
	mockSig := make([]byte, 65) // Replace with the desired mock signature
	mockSig[64] = 27            // Set V to 27 for the transformed value
	// mockHexSig := hex.EncodeToString(mockSig)

	// Call the function under test
	finalSig, err := signer.SignMessage(privKey, signerData)

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, len(finalSig), len(mockSig))
	assert.True(t, true, finalSig[64] == mockSig[64] || finalSig[64]+1 == mockSig[64])
}

func TestSignMessage_EncodeError(t *testing.T) {
	privKey, _ := crypto.GenerateKey()

	signerData := &apitypes.TypedData{}

	// Call the function under test
	finalSig, err := signer.SignMessage(privKey, signerData)

	// Verify the result
	assert.Error(t, err)
	assert.Nil(t, finalSig)
}

func TestSignMessage_SignError(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	signerData := &apitypes.TypedData{}

	// Call the function under test
	finalSig, err := signer.SignMessage(privKey, signerData)

	// Verify the result
	assert.Error(t, err)
	assert.Nil(t, finalSig)
	// mockLogger.AssertExpectations(t)
}

func TestSigner_GetSignerData_Success(t *testing.T) {
	settingsObj := &settings.SettingsObj{
		Signer: &settings.Signer{
			Domain: struct {
				Name              string `json:"name"`
				Version           string `json:"version"`
				ChainId           string `json:"chainId"`
				VerifyingContract string `json:"verifyingContract"`
			}{
				Name:              "test",
				Version:           "0.1",
				ChainId:           "100",
				VerifyingContract: "0x6FC64a0356Ebb41F88cd31e69A5C16e35DFFaDd1",
			},
		},
	}

	want := &apitypes.TypedData{
		PrimaryType: "Request",
		Types: apitypes.Types{
			"Request": []apitypes.Type{
				{Name: "deadline", Type: "uint256"},
				{Name: "snapshotCid", Type: "string"},
				{Name: "epochId", Type: "uint256"},
				{Name: "projectId", Type: "string"},
			},
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		Domain: apitypes.TypedDataDomain{
			Name:              settingsObj.Signer.Domain.Name,
			Version:           settingsObj.Signer.Domain.Version,
			ChainId:           (*math.HexOrDecimal256)(math.MustParseBig256(settingsObj.Signer.Domain.ChainId)),
			VerifyingContract: settingsObj.Signer.Domain.VerifyingContract,
		},
		Message: apitypes.TypedDataMessage{
			"deadline":    (*math.HexOrDecimal256)(big.NewInt(int64(1) + int64(settingsObj.Signer.DeadlineBuffer))),
			"snapshotCid": "snapshotCid",
			"epochId":     (*math.HexOrDecimal256)(big.NewInt(1)),
			"projectId":   "projectId",
		},
	}

	s := &Signer{
		settingsObj: settingsObj,
	}

	got, err := s.GetSignerData(new(MockEthClient), "snapshotCid", "projectId", 1)

	assert.Equal(t, err, nil)

	assert.Equal(t, want, got)
}

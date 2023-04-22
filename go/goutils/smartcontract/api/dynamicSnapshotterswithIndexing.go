// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractApi

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// AuditRecordStoreDynamicSnapshottersWithIndexingRequest is an auto generated low-level Go binding around an user-defined struct.
type AuditRecordStoreDynamicSnapshottersWithIndexingRequest struct {
	Deadline *big.Int
}

// ContractApiMetaData contains all meta data concerning the ContractApi contract.
var ContractApiMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"adminModuleAddress\",\"type\":\"address\"}],\"name\":\"AdminModuleUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"dagCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"AggregateDagCidFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"dagCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"AggregateDagCidSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"aggregateCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"AggregateFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"aggregateCid\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"AggregateSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"dagCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DagCidFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"aggregateCid\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DelayedAggregateSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"indexTailDAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"indexIdentifierHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DelayedIndexSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DelayedSnapshotSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"EpochReleased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"indexTailDAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tailBlockEpochSourceChainHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"indexIdentifierHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"IndexFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"indexTailDAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"indexIdentifierHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"IndexSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"dagCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotDagCidFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"dagCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotDagCidSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"}],\"name\":\"SnapshotterAllowed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"}],\"name\":\"SnapshotterRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"stateBuilderAddress\",\"type\":\"address\"}],\"name\":\"StateBuilderAllowed\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operatorAddr\",\"type\":\"address\"}],\"name\":\"addOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"aggregateReceived\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"aggregateReceivedCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"aggregateStartTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blocknumber\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"aggregateStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"finalized\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"aggregateSubmissionWindow\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"}],\"name\":\"allowSnapshotter\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"}],\"name\":\"checkDynamicConsensusAggregate\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"indexIdentifierHash\",\"type\":\"bytes32\"}],\"name\":\"checkDynamicConsensusIndex\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"}],\"name\":\"checkDynamicConsensusSnapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"dagBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"dagCid\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structAuditRecordStoreDynamicSnapshottersWithIndexing.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"commitFinalizedDAGcid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"epochReleaseTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blocknumber\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"finalizedDagCids\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"finalizedIndexes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tailDAGBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"tailDAGBlockEpochSourceChainHeight\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"}],\"name\":\"forceCompleteConsensusAggregate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"indexIdentifierHash\",\"type\":\"bytes32\"}],\"name\":\"forceCompleteConsensusIndex\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"}],\"name\":\"forceCompleteConsensusSnapshot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllSnapshotters\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOperators\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProjects\",\"outputs\":[{\"internalType\":\"string[]\",\"name\":\"\",\"type\":\"string[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalSnapshotterCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"indexReceived\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"indexReceivedCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"indexStartTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blocknumber\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"indexStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"finalized\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"indexSubmissionWindow\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxAggregatesCid\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxAggregatesCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxIndexCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxIndexData\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tailDAGBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"tailDAGBlockEpochSourceChainHeight\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxSnapshotsCid\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxSnapshotsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minSubmissionsForConsensus\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"projectFirstEpochEndHeight\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"messageHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"recoverAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"releaseEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operatorAddr\",\"type\":\"address\"}],\"name\":\"removeOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"snapshotStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"finalized\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"snapshotSubmissionWindow\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"snapshotsReceived\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"snapshotsReceivedCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"snapshotters\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structAuditRecordStoreDynamicSnapshottersWithIndexing.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"submitAggregate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"DAGBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"indexTailDAGBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"tailBlockEpochSourceChainHeight\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"indexIdentifierHash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structAuditRecordStoreDynamicSnapshottersWithIndexing.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"submitIndex\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structAuditRecordStoreDynamicSnapshottersWithIndexing.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"submitSnapshot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSnapshotterCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newAggregateSubmissionWindow\",\"type\":\"uint256\"}],\"name\":\"updateAggregateSubmissionWindow\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newIndexSubmissionWindow\",\"type\":\"uint256\"}],\"name\":\"updateIndexSubmissionWindow\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minSubmissionsForConsensus\",\"type\":\"uint256\"}],\"name\":\"updateMinSnapshottersForConsensus\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newsnapshotSubmissionWindow\",\"type\":\"uint256\"}],\"name\":\"updateSnapshotSubmissionWindow\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"internalType\":\"structAuditRecordStoreDynamicSnapshottersWithIndexing.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"verify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5060405161096438038061096483398101604081905261002f916100e2565b600080546001600160a01b03191633908117825581526001602081905260408220555b81518110156100c5576002604051806040016040528084848151811061007a5761007a61019f565b602090810291909101810151825260009181018290528354600181810186559483529181902083516002909302019182559190910151910155806100bd816101b5565b915050610052565b50506101dc565b634e487b7160e01b600052604160045260246000fd5b600060208083850312156100f557600080fd5b82516001600160401b038082111561010c57600080fd5b818501915085601f83011261012057600080fd5b815181811115610132576101326100cc565b8060051b604051601f19603f83011681018181108582111715610157576101576100cc565b60405291825284820192508381018501918883111561017557600080fd5b938501935b828510156101935784518452938501939285019261017a565b98975050505050505050565b634e487b7160e01b600052603260045260246000fd5b6000600182016101d557634e487b7160e01b600052601160045260246000fd5b5060010190565b610779806101eb6000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063609ff1bd1161005b578063609ff1bd1461010d5780639e7b8d6114610123578063a3ec138d14610136578063e2ba53f0146101a757600080fd5b80630121b93f1461008d578063013cf08b146100a25780632e4176cf146100cf5780635c19a95c146100fa575b600080fd5b6100a061009b36600461069c565b6101af565b005b6100b56100b036600461069c565b6102a9565b604080519283526020830191909152015b60405180910390f35b6000546100e2906001600160a01b031681565b6040516001600160a01b0390911681526020016100c6565b6100a06101083660046106b5565b6102d7565b6101156104d4565b6040519081526020016100c6565b6100a06101313660046106b5565b610551565b6101786101443660046106b5565b600160208190526000918252604090912080549181015460029091015460ff82169161010090046001600160a01b03169084565b6040516100c6949392919093845291151560208401526001600160a01b03166040830152606082015260800190565b610115610669565b336000908152600160205260408120805490910361020b5760405162461bcd60e51b8152602060048201526014602482015273486173206e6f20726967687420746f20766f746560601b60448201526064015b60405180910390fd5b600181015460ff16156102515760405162461bcd60e51b815260206004820152600e60248201526d20b63932b0b23c903b37ba32b21760911b6044820152606401610202565b6001818101805460ff1916909117905560028082018390558154815490919084908110610280576102806106e5565b906000526020600020906002020160010160008282546102a09190610711565b90915550505050565b600281815481106102b957600080fd5b60009182526020909120600290910201805460019091015490915082565b3360009081526001602081905260409091209081015460ff16156103325760405162461bcd60e51b81526020600482015260126024820152712cb7ba9030b63932b0b23c903b37ba32b21760711b6044820152606401610202565b336001600160a01b0383160361038a5760405162461bcd60e51b815260206004820152601e60248201527f53656c662d64656c65676174696f6e20697320646973616c6c6f7765642e00006044820152606401610202565b6001600160a01b03828116600090815260016020819052604090912001546101009004161561042e576001600160a01b03918216600090815260016020819052604090912001546101009004909116903382036104295760405162461bcd60e51b815260206004820152601960248201527f466f756e64206c6f6f7020696e2064656c65676174696f6e2e000000000000006044820152606401610202565b61038a565b600181810180546001600160a81b0319166101006001600160a01b03861690810291909117831790915560009081526020829052604090209081015460ff16156104b55781546002828101548154811061048a5761048a6106e5565b906000526020600020906002020160010160008282546104aa9190610711565b909155506104cf9050565b8154815482906000906104c9908490610711565b90915550505b505050565b600080805b60025481101561054c5781600282815481106104f7576104f76106e5565b906000526020600020906002020160010154111561053a5760028181548110610522576105226106e5565b90600052602060002090600202016001015491508092505b806105448161072a565b9150506104d9565b505090565b6000546001600160a01b031633146105bc5760405162461bcd60e51b815260206004820152602860248201527f4f6e6c79206368616972706572736f6e2063616e2067697665207269676874206044820152673a37903b37ba329760c11b6064820152608401610202565b6001600160a01b0381166000908152600160208190526040909120015460ff16156106295760405162461bcd60e51b815260206004820152601860248201527f54686520766f74657220616c726561647920766f7465642e00000000000000006044820152606401610202565b6001600160a01b0381166000908152600160205260409020541561064c57600080fd5b6001600160a01b0316600090815260016020819052604090912055565b600060026106756104d4565b81548110610685576106856106e5565b906000526020600020906002020160000154905090565b6000602082840312156106ae57600080fd5b5035919050565b6000602082840312156106c757600080fd5b81356001600160a01b03811681146106de57600080fd5b9392505050565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052601160045260246000fd5b80820180821115610724576107246106fb565b92915050565b60006001820161073c5761073c6106fb565b506001019056fea26469706673582212204993ffc89fa7ef29c373cf18a3d8b12fa1a856ecd144c288a1f460ea0d665bc064736f6c63430008110033",
}

// ContractApiABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractApiMetaData.ABI instead.
var ContractApiABI = ContractApiMetaData.ABI

// ContractApiBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractApiMetaData.Bin instead.
var ContractApiBin = ContractApiMetaData.Bin

// DeployContractApi deploys a new Ethereum contract, binding an instance of ContractApi to it.
func DeployContractApi(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ContractApi, error) {
	parsed, err := ContractApiMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractApiBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractApi{ContractApiCaller: ContractApiCaller{contract: contract}, ContractApiTransactor: ContractApiTransactor{contract: contract}, ContractApiFilterer: ContractApiFilterer{contract: contract}}, nil
}

// ContractApi is an auto generated Go binding around an Ethereum contract.
type ContractApi struct {
	ContractApiCaller     // Read-only binding to the contract
	ContractApiTransactor // Write-only binding to the contract
	ContractApiFilterer   // Log filterer for contract events
}

// ContractApiCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractApiCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractApiTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractApiTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractApiFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractApiFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractApiSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractApiSession struct {
	Contract     *ContractApi      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractApiCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractApiCallerSession struct {
	Contract *ContractApiCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// ContractApiTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractApiTransactorSession struct {
	Contract     *ContractApiTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// ContractApiRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractApiRaw struct {
	Contract *ContractApi // Generic contract binding to access the raw methods on
}

// ContractApiCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractApiCallerRaw struct {
	Contract *ContractApiCaller // Generic read-only contract binding to access the raw methods on
}

// ContractApiTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractApiTransactorRaw struct {
	Contract *ContractApiTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractApi creates a new instance of ContractApi, bound to a specific deployed contract.
func NewContractApi(address common.Address, backend bind.ContractBackend) (*ContractApi, error) {
	contract, err := bindContractApi(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractApi{ContractApiCaller: ContractApiCaller{contract: contract}, ContractApiTransactor: ContractApiTransactor{contract: contract}, ContractApiFilterer: ContractApiFilterer{contract: contract}}, nil
}

// NewContractApiCaller creates a new read-only instance of ContractApi, bound to a specific deployed contract.
func NewContractApiCaller(address common.Address, caller bind.ContractCaller) (*ContractApiCaller, error) {
	contract, err := bindContractApi(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractApiCaller{contract: contract}, nil
}

// NewContractApiTransactor creates a new write-only instance of ContractApi, bound to a specific deployed contract.
func NewContractApiTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractApiTransactor, error) {
	contract, err := bindContractApi(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractApiTransactor{contract: contract}, nil
}

// NewContractApiFilterer creates a new log filterer instance of ContractApi, bound to a specific deployed contract.
func NewContractApiFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractApiFilterer, error) {
	contract, err := bindContractApi(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractApiFilterer{contract: contract}, nil
}

// bindContractApi binds a generic wrapper to an already deployed contract.
func bindContractApi(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractApiMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractApi *ContractApiRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractApi.Contract.ContractApiCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractApi *ContractApiRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractApi.Contract.ContractApiTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractApi *ContractApiRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractApi.Contract.ContractApiTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractApi *ContractApiCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractApi.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractApi *ContractApiTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractApi.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractApi *ContractApiTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractApi.Contract.contract.Transact(opts, method, params...)
}

// AggregateReceived is a free data retrieval call binding the contract method 0x5cec9057.
//
// Solidity: function aggregateReceived(string , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiCaller) AggregateReceived(opts *bind.CallOpts, arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "aggregateReceived", arg0, arg1, arg2)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AggregateReceived is a free data retrieval call binding the contract method 0x5cec9057.
//
// Solidity: function aggregateReceived(string , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiSession) AggregateReceived(arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	return _ContractApi.Contract.AggregateReceived(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// AggregateReceived is a free data retrieval call binding the contract method 0x5cec9057.
//
// Solidity: function aggregateReceived(string , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiCallerSession) AggregateReceived(arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	return _ContractApi.Contract.AggregateReceived(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// AggregateReceivedCount is a free data retrieval call binding the contract method 0xca1cd46e.
//
// Solidity: function aggregateReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ContractApi *ContractApiCaller) AggregateReceivedCount(opts *bind.CallOpts, arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "aggregateReceivedCount", arg0, arg1, arg2)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AggregateReceivedCount is a free data retrieval call binding the contract method 0xca1cd46e.
//
// Solidity: function aggregateReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ContractApi *ContractApiSession) AggregateReceivedCount(arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	return _ContractApi.Contract.AggregateReceivedCount(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// AggregateReceivedCount is a free data retrieval call binding the contract method 0xca1cd46e.
//
// Solidity: function aggregateReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) AggregateReceivedCount(arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	return _ContractApi.Contract.AggregateReceivedCount(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// AggregateStartTime is a free data retrieval call binding the contract method 0x817cc985.
//
// Solidity: function aggregateStartTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiCaller) AggregateStartTime(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "aggregateStartTime", arg0)

	outstruct := new(struct {
		Timestamp   *big.Int
		Blocknumber *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Timestamp = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Blocknumber = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// AggregateStartTime is a free data retrieval call binding the contract method 0x817cc985.
//
// Solidity: function aggregateStartTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiSession) AggregateStartTime(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	return _ContractApi.Contract.AggregateStartTime(&_ContractApi.CallOpts, arg0)
}

// AggregateStartTime is a free data retrieval call binding the contract method 0x817cc985.
//
// Solidity: function aggregateStartTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiCallerSession) AggregateStartTime(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	return _ContractApi.Contract.AggregateStartTime(&_ContractApi.CallOpts, arg0)
}

// AggregateStatus is a free data retrieval call binding the contract method 0x23ec78ec.
//
// Solidity: function aggregateStatus(string , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiCaller) AggregateStatus(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "aggregateStatus", arg0, arg1)

	outstruct := new(struct {
		Finalized bool
		Timestamp *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Finalized = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Timestamp = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// AggregateStatus is a free data retrieval call binding the contract method 0x23ec78ec.
//
// Solidity: function aggregateStatus(string , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiSession) AggregateStatus(arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	return _ContractApi.Contract.AggregateStatus(&_ContractApi.CallOpts, arg0, arg1)
}

// AggregateStatus is a free data retrieval call binding the contract method 0x23ec78ec.
//
// Solidity: function aggregateStatus(string , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiCallerSession) AggregateStatus(arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	return _ContractApi.Contract.AggregateStatus(&_ContractApi.CallOpts, arg0, arg1)
}

// AggregateSubmissionWindow is a free data retrieval call binding the contract method 0x2207f05c.
//
// Solidity: function aggregateSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiCaller) AggregateSubmissionWindow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "aggregateSubmissionWindow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AggregateSubmissionWindow is a free data retrieval call binding the contract method 0x2207f05c.
//
// Solidity: function aggregateSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiSession) AggregateSubmissionWindow() (*big.Int, error) {
	return _ContractApi.Contract.AggregateSubmissionWindow(&_ContractApi.CallOpts)
}

// AggregateSubmissionWindow is a free data retrieval call binding the contract method 0x2207f05c.
//
// Solidity: function aggregateSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) AggregateSubmissionWindow() (*big.Int, error) {
	return _ContractApi.Contract.AggregateSubmissionWindow(&_ContractApi.CallOpts)
}

// CheckDynamicConsensusAggregate is a free data retrieval call binding the contract method 0xf2b3fe91.
//
// Solidity: function checkDynamicConsensusAggregate(string projectId, uint256 DAGBlockHeight) view returns(bool success)
func (_ContractApi *ContractApiCaller) CheckDynamicConsensusAggregate(opts *bind.CallOpts, projectId string, DAGBlockHeight *big.Int) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "checkDynamicConsensusAggregate", projectId, DAGBlockHeight)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckDynamicConsensusAggregate is a free data retrieval call binding the contract method 0xf2b3fe91.
//
// Solidity: function checkDynamicConsensusAggregate(string projectId, uint256 DAGBlockHeight) view returns(bool success)
func (_ContractApi *ContractApiSession) CheckDynamicConsensusAggregate(projectId string, DAGBlockHeight *big.Int) (bool, error) {
	return _ContractApi.Contract.CheckDynamicConsensusAggregate(&_ContractApi.CallOpts, projectId, DAGBlockHeight)
}

// CheckDynamicConsensusAggregate is a free data retrieval call binding the contract method 0xf2b3fe91.
//
// Solidity: function checkDynamicConsensusAggregate(string projectId, uint256 DAGBlockHeight) view returns(bool success)
func (_ContractApi *ContractApiCallerSession) CheckDynamicConsensusAggregate(projectId string, DAGBlockHeight *big.Int) (bool, error) {
	return _ContractApi.Contract.CheckDynamicConsensusAggregate(&_ContractApi.CallOpts, projectId, DAGBlockHeight)
}

// CheckDynamicConsensusIndex is a free data retrieval call binding the contract method 0x257bb0a4.
//
// Solidity: function checkDynamicConsensusIndex(string projectId, uint256 DAGBlockHeight, bytes32 indexIdentifierHash) view returns(bool success)
func (_ContractApi *ContractApiCaller) CheckDynamicConsensusIndex(opts *bind.CallOpts, projectId string, DAGBlockHeight *big.Int, indexIdentifierHash [32]byte) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "checkDynamicConsensusIndex", projectId, DAGBlockHeight, indexIdentifierHash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckDynamicConsensusIndex is a free data retrieval call binding the contract method 0x257bb0a4.
//
// Solidity: function checkDynamicConsensusIndex(string projectId, uint256 DAGBlockHeight, bytes32 indexIdentifierHash) view returns(bool success)
func (_ContractApi *ContractApiSession) CheckDynamicConsensusIndex(projectId string, DAGBlockHeight *big.Int, indexIdentifierHash [32]byte) (bool, error) {
	return _ContractApi.Contract.CheckDynamicConsensusIndex(&_ContractApi.CallOpts, projectId, DAGBlockHeight, indexIdentifierHash)
}

// CheckDynamicConsensusIndex is a free data retrieval call binding the contract method 0x257bb0a4.
//
// Solidity: function checkDynamicConsensusIndex(string projectId, uint256 DAGBlockHeight, bytes32 indexIdentifierHash) view returns(bool success)
func (_ContractApi *ContractApiCallerSession) CheckDynamicConsensusIndex(projectId string, DAGBlockHeight *big.Int, indexIdentifierHash [32]byte) (bool, error) {
	return _ContractApi.Contract.CheckDynamicConsensusIndex(&_ContractApi.CallOpts, projectId, DAGBlockHeight, indexIdentifierHash)
}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochEnd) view returns(bool success)
func (_ContractApi *ContractApiCaller) CheckDynamicConsensusSnapshot(opts *bind.CallOpts, projectId string, epochEnd *big.Int) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "checkDynamicConsensusSnapshot", projectId, epochEnd)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochEnd) view returns(bool success)
func (_ContractApi *ContractApiSession) CheckDynamicConsensusSnapshot(projectId string, epochEnd *big.Int) (bool, error) {
	return _ContractApi.Contract.CheckDynamicConsensusSnapshot(&_ContractApi.CallOpts, projectId, epochEnd)
}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochEnd) view returns(bool success)
func (_ContractApi *ContractApiCallerSession) CheckDynamicConsensusSnapshot(projectId string, epochEnd *big.Int) (bool, error) {
	return _ContractApi.Contract.CheckDynamicConsensusSnapshot(&_ContractApi.CallOpts, projectId, epochEnd)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256 begin, uint256 end)
func (_ContractApi *ContractApiCaller) CurrentEpoch(opts *bind.CallOpts) (struct {
	Begin *big.Int
	End   *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "currentEpoch")

	outstruct := new(struct {
		Begin *big.Int
		End   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Begin = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.End = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256 begin, uint256 end)
func (_ContractApi *ContractApiSession) CurrentEpoch() (struct {
	Begin *big.Int
	End   *big.Int
}, error) {
	return _ContractApi.Contract.CurrentEpoch(&_ContractApi.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256 begin, uint256 end)
func (_ContractApi *ContractApiCallerSession) CurrentEpoch() (struct {
	Begin *big.Int
	End   *big.Int
}, error) {
	return _ContractApi.Contract.CurrentEpoch(&_ContractApi.CallOpts)
}

// EpochReleaseTime is a free data retrieval call binding the contract method 0x977391a9.
//
// Solidity: function epochReleaseTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiCaller) EpochReleaseTime(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "epochReleaseTime", arg0)

	outstruct := new(struct {
		Timestamp   *big.Int
		Blocknumber *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Timestamp = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Blocknumber = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// EpochReleaseTime is a free data retrieval call binding the contract method 0x977391a9.
//
// Solidity: function epochReleaseTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiSession) EpochReleaseTime(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	return _ContractApi.Contract.EpochReleaseTime(&_ContractApi.CallOpts, arg0)
}

// EpochReleaseTime is a free data retrieval call binding the contract method 0x977391a9.
//
// Solidity: function epochReleaseTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiCallerSession) EpochReleaseTime(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	return _ContractApi.Contract.EpochReleaseTime(&_ContractApi.CallOpts, arg0)
}

// FinalizedDagCids is a free data retrieval call binding the contract method 0x107aa603.
//
// Solidity: function finalizedDagCids(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiCaller) FinalizedDagCids(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (string, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "finalizedDagCids", arg0, arg1)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// FinalizedDagCids is a free data retrieval call binding the contract method 0x107aa603.
//
// Solidity: function finalizedDagCids(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiSession) FinalizedDagCids(arg0 string, arg1 *big.Int) (string, error) {
	return _ContractApi.Contract.FinalizedDagCids(&_ContractApi.CallOpts, arg0, arg1)
}

// FinalizedDagCids is a free data retrieval call binding the contract method 0x107aa603.
//
// Solidity: function finalizedDagCids(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiCallerSession) FinalizedDagCids(arg0 string, arg1 *big.Int) (string, error) {
	return _ContractApi.Contract.FinalizedDagCids(&_ContractApi.CallOpts, arg0, arg1)
}

// FinalizedIndexes is a free data retrieval call binding the contract method 0xe26926f2.
//
// Solidity: function finalizedIndexes(string , bytes32 , uint256 ) view returns(uint256 tailDAGBlockHeight, uint256 tailDAGBlockEpochSourceChainHeight)
func (_ContractApi *ContractApiCaller) FinalizedIndexes(opts *bind.CallOpts, arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	TailDAGBlockHeight                 *big.Int
	TailDAGBlockEpochSourceChainHeight *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "finalizedIndexes", arg0, arg1, arg2)

	outstruct := new(struct {
		TailDAGBlockHeight                 *big.Int
		TailDAGBlockEpochSourceChainHeight *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TailDAGBlockHeight = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TailDAGBlockEpochSourceChainHeight = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// FinalizedIndexes is a free data retrieval call binding the contract method 0xe26926f2.
//
// Solidity: function finalizedIndexes(string , bytes32 , uint256 ) view returns(uint256 tailDAGBlockHeight, uint256 tailDAGBlockEpochSourceChainHeight)
func (_ContractApi *ContractApiSession) FinalizedIndexes(arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	TailDAGBlockHeight                 *big.Int
	TailDAGBlockEpochSourceChainHeight *big.Int
}, error) {
	return _ContractApi.Contract.FinalizedIndexes(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// FinalizedIndexes is a free data retrieval call binding the contract method 0xe26926f2.
//
// Solidity: function finalizedIndexes(string , bytes32 , uint256 ) view returns(uint256 tailDAGBlockHeight, uint256 tailDAGBlockEpochSourceChainHeight)
func (_ContractApi *ContractApiCallerSession) FinalizedIndexes(arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	TailDAGBlockHeight                 *big.Int
	TailDAGBlockEpochSourceChainHeight *big.Int
}, error) {
	return _ContractApi.Contract.FinalizedIndexes(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// GetAllSnapshotters is a free data retrieval call binding the contract method 0x38f3ce5f.
//
// Solidity: function getAllSnapshotters() view returns(address[])
func (_ContractApi *ContractApiCaller) GetAllSnapshotters(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getAllSnapshotters")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetAllSnapshotters is a free data retrieval call binding the contract method 0x38f3ce5f.
//
// Solidity: function getAllSnapshotters() view returns(address[])
func (_ContractApi *ContractApiSession) GetAllSnapshotters() ([]common.Address, error) {
	return _ContractApi.Contract.GetAllSnapshotters(&_ContractApi.CallOpts)
}

// GetAllSnapshotters is a free data retrieval call binding the contract method 0x38f3ce5f.
//
// Solidity: function getAllSnapshotters() view returns(address[])
func (_ContractApi *ContractApiCallerSession) GetAllSnapshotters() ([]common.Address, error) {
	return _ContractApi.Contract.GetAllSnapshotters(&_ContractApi.CallOpts)
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() view returns(address[])
func (_ContractApi *ContractApiCaller) GetOperators(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getOperators")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() view returns(address[])
func (_ContractApi *ContractApiSession) GetOperators() ([]common.Address, error) {
	return _ContractApi.Contract.GetOperators(&_ContractApi.CallOpts)
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() view returns(address[])
func (_ContractApi *ContractApiCallerSession) GetOperators() ([]common.Address, error) {
	return _ContractApi.Contract.GetOperators(&_ContractApi.CallOpts)
}

// GetProjects is a free data retrieval call binding the contract method 0xdcc60128.
//
// Solidity: function getProjects() view returns(string[])
func (_ContractApi *ContractApiCaller) GetProjects(opts *bind.CallOpts) ([]string, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getProjects")

	if err != nil {
		return *new([]string), err
	}

	out0 := *abi.ConvertType(out[0], new([]string)).(*[]string)

	return out0, err

}

// GetProjects is a free data retrieval call binding the contract method 0xdcc60128.
//
// Solidity: function getProjects() view returns(string[])
func (_ContractApi *ContractApiSession) GetProjects() ([]string, error) {
	return _ContractApi.Contract.GetProjects(&_ContractApi.CallOpts)
}

// GetProjects is a free data retrieval call binding the contract method 0xdcc60128.
//
// Solidity: function getProjects() view returns(string[])
func (_ContractApi *ContractApiCallerSession) GetProjects() ([]string, error) {
	return _ContractApi.Contract.GetProjects(&_ContractApi.CallOpts)
}

// GetTotalSnapshotterCount is a free data retrieval call binding the contract method 0x92ae6f66.
//
// Solidity: function getTotalSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiCaller) GetTotalSnapshotterCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getTotalSnapshotterCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalSnapshotterCount is a free data retrieval call binding the contract method 0x92ae6f66.
//
// Solidity: function getTotalSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiSession) GetTotalSnapshotterCount() (*big.Int, error) {
	return _ContractApi.Contract.GetTotalSnapshotterCount(&_ContractApi.CallOpts)
}

// GetTotalSnapshotterCount is a free data retrieval call binding the contract method 0x92ae6f66.
//
// Solidity: function getTotalSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) GetTotalSnapshotterCount() (*big.Int, error) {
	return _ContractApi.Contract.GetTotalSnapshotterCount(&_ContractApi.CallOpts)
}

// IndexReceived is a free data retrieval call binding the contract method 0x3c0d5981.
//
// Solidity: function indexReceived(string , bytes32 , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiCaller) IndexReceived(opts *bind.CallOpts, arg0 string, arg1 [32]byte, arg2 *big.Int, arg3 common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "indexReceived", arg0, arg1, arg2, arg3)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IndexReceived is a free data retrieval call binding the contract method 0x3c0d5981.
//
// Solidity: function indexReceived(string , bytes32 , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiSession) IndexReceived(arg0 string, arg1 [32]byte, arg2 *big.Int, arg3 common.Address) (bool, error) {
	return _ContractApi.Contract.IndexReceived(&_ContractApi.CallOpts, arg0, arg1, arg2, arg3)
}

// IndexReceived is a free data retrieval call binding the contract method 0x3c0d5981.
//
// Solidity: function indexReceived(string , bytes32 , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiCallerSession) IndexReceived(arg0 string, arg1 [32]byte, arg2 *big.Int, arg3 common.Address) (bool, error) {
	return _ContractApi.Contract.IndexReceived(&_ContractApi.CallOpts, arg0, arg1, arg2, arg3)
}

// IndexReceivedCount is a free data retrieval call binding the contract method 0x4af2e85e.
//
// Solidity: function indexReceivedCount(string , bytes32 , uint256 , bytes32 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) IndexReceivedCount(opts *bind.CallOpts, arg0 string, arg1 [32]byte, arg2 *big.Int, arg3 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "indexReceivedCount", arg0, arg1, arg2, arg3)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// IndexReceivedCount is a free data retrieval call binding the contract method 0x4af2e85e.
//
// Solidity: function indexReceivedCount(string , bytes32 , uint256 , bytes32 ) view returns(uint256)
func (_ContractApi *ContractApiSession) IndexReceivedCount(arg0 string, arg1 [32]byte, arg2 *big.Int, arg3 [32]byte) (*big.Int, error) {
	return _ContractApi.Contract.IndexReceivedCount(&_ContractApi.CallOpts, arg0, arg1, arg2, arg3)
}

// IndexReceivedCount is a free data retrieval call binding the contract method 0x4af2e85e.
//
// Solidity: function indexReceivedCount(string , bytes32 , uint256 , bytes32 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) IndexReceivedCount(arg0 string, arg1 [32]byte, arg2 *big.Int, arg3 [32]byte) (*big.Int, error) {
	return _ContractApi.Contract.IndexReceivedCount(&_ContractApi.CallOpts, arg0, arg1, arg2, arg3)
}

// IndexStartTime is a free data retrieval call binding the contract method 0x08282030.
//
// Solidity: function indexStartTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiCaller) IndexStartTime(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "indexStartTime", arg0)

	outstruct := new(struct {
		Timestamp   *big.Int
		Blocknumber *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Timestamp = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Blocknumber = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// IndexStartTime is a free data retrieval call binding the contract method 0x08282030.
//
// Solidity: function indexStartTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiSession) IndexStartTime(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	return _ContractApi.Contract.IndexStartTime(&_ContractApi.CallOpts, arg0)
}

// IndexStartTime is a free data retrieval call binding the contract method 0x08282030.
//
// Solidity: function indexStartTime(uint256 ) view returns(uint256 timestamp, uint256 blocknumber)
func (_ContractApi *ContractApiCallerSession) IndexStartTime(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
}, error) {
	return _ContractApi.Contract.IndexStartTime(&_ContractApi.CallOpts, arg0)
}

// IndexStatus is a free data retrieval call binding the contract method 0x7cd6b463.
//
// Solidity: function indexStatus(string , bytes32 , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiCaller) IndexStatus(opts *bind.CallOpts, arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "indexStatus", arg0, arg1, arg2)

	outstruct := new(struct {
		Finalized bool
		Timestamp *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Finalized = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Timestamp = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// IndexStatus is a free data retrieval call binding the contract method 0x7cd6b463.
//
// Solidity: function indexStatus(string , bytes32 , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiSession) IndexStatus(arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	return _ContractApi.Contract.IndexStatus(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// IndexStatus is a free data retrieval call binding the contract method 0x7cd6b463.
//
// Solidity: function indexStatus(string , bytes32 , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiCallerSession) IndexStatus(arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	return _ContractApi.Contract.IndexStatus(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// IndexSubmissionWindow is a free data retrieval call binding the contract method 0x390c27a9.
//
// Solidity: function indexSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiCaller) IndexSubmissionWindow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "indexSubmissionWindow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// IndexSubmissionWindow is a free data retrieval call binding the contract method 0x390c27a9.
//
// Solidity: function indexSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiSession) IndexSubmissionWindow() (*big.Int, error) {
	return _ContractApi.Contract.IndexSubmissionWindow(&_ContractApi.CallOpts)
}

// IndexSubmissionWindow is a free data retrieval call binding the contract method 0x390c27a9.
//
// Solidity: function indexSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) IndexSubmissionWindow() (*big.Int, error) {
	return _ContractApi.Contract.IndexSubmissionWindow(&_ContractApi.CallOpts)
}

// MaxAggregatesCid is a free data retrieval call binding the contract method 0x933421e3.
//
// Solidity: function maxAggregatesCid(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiCaller) MaxAggregatesCid(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (string, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "maxAggregatesCid", arg0, arg1)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// MaxAggregatesCid is a free data retrieval call binding the contract method 0x933421e3.
//
// Solidity: function maxAggregatesCid(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiSession) MaxAggregatesCid(arg0 string, arg1 *big.Int) (string, error) {
	return _ContractApi.Contract.MaxAggregatesCid(&_ContractApi.CallOpts, arg0, arg1)
}

// MaxAggregatesCid is a free data retrieval call binding the contract method 0x933421e3.
//
// Solidity: function maxAggregatesCid(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiCallerSession) MaxAggregatesCid(arg0 string, arg1 *big.Int) (string, error) {
	return _ContractApi.Contract.MaxAggregatesCid(&_ContractApi.CallOpts, arg0, arg1)
}

// MaxAggregatesCount is a free data retrieval call binding the contract method 0xeba521c5.
//
// Solidity: function maxAggregatesCount(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) MaxAggregatesCount(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "maxAggregatesCount", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxAggregatesCount is a free data retrieval call binding the contract method 0xeba521c5.
//
// Solidity: function maxAggregatesCount(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) MaxAggregatesCount(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.MaxAggregatesCount(&_ContractApi.CallOpts, arg0, arg1)
}

// MaxAggregatesCount is a free data retrieval call binding the contract method 0xeba521c5.
//
// Solidity: function maxAggregatesCount(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) MaxAggregatesCount(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.MaxAggregatesCount(&_ContractApi.CallOpts, arg0, arg1)
}

// MaxIndexCount is a free data retrieval call binding the contract method 0x1c5ddfd0.
//
// Solidity: function maxIndexCount(string , bytes32 , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) MaxIndexCount(opts *bind.CallOpts, arg0 string, arg1 [32]byte, arg2 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "maxIndexCount", arg0, arg1, arg2)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxIndexCount is a free data retrieval call binding the contract method 0x1c5ddfd0.
//
// Solidity: function maxIndexCount(string , bytes32 , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) MaxIndexCount(arg0 string, arg1 [32]byte, arg2 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.MaxIndexCount(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// MaxIndexCount is a free data retrieval call binding the contract method 0x1c5ddfd0.
//
// Solidity: function maxIndexCount(string , bytes32 , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) MaxIndexCount(arg0 string, arg1 [32]byte, arg2 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.MaxIndexCount(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// MaxIndexData is a free data retrieval call binding the contract method 0x2b583dc4.
//
// Solidity: function maxIndexData(string , bytes32 , uint256 ) view returns(uint256 tailDAGBlockHeight, uint256 tailDAGBlockEpochSourceChainHeight)
func (_ContractApi *ContractApiCaller) MaxIndexData(opts *bind.CallOpts, arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	TailDAGBlockHeight                 *big.Int
	TailDAGBlockEpochSourceChainHeight *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "maxIndexData", arg0, arg1, arg2)

	outstruct := new(struct {
		TailDAGBlockHeight                 *big.Int
		TailDAGBlockEpochSourceChainHeight *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TailDAGBlockHeight = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TailDAGBlockEpochSourceChainHeight = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// MaxIndexData is a free data retrieval call binding the contract method 0x2b583dc4.
//
// Solidity: function maxIndexData(string , bytes32 , uint256 ) view returns(uint256 tailDAGBlockHeight, uint256 tailDAGBlockEpochSourceChainHeight)
func (_ContractApi *ContractApiSession) MaxIndexData(arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	TailDAGBlockHeight                 *big.Int
	TailDAGBlockEpochSourceChainHeight *big.Int
}, error) {
	return _ContractApi.Contract.MaxIndexData(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// MaxIndexData is a free data retrieval call binding the contract method 0x2b583dc4.
//
// Solidity: function maxIndexData(string , bytes32 , uint256 ) view returns(uint256 tailDAGBlockHeight, uint256 tailDAGBlockEpochSourceChainHeight)
func (_ContractApi *ContractApiCallerSession) MaxIndexData(arg0 string, arg1 [32]byte, arg2 *big.Int) (struct {
	TailDAGBlockHeight                 *big.Int
	TailDAGBlockEpochSourceChainHeight *big.Int
}, error) {
	return _ContractApi.Contract.MaxIndexData(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// MaxSnapshotsCid is a free data retrieval call binding the contract method 0xc2b97d4c.
//
// Solidity: function maxSnapshotsCid(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiCaller) MaxSnapshotsCid(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (string, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "maxSnapshotsCid", arg0, arg1)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// MaxSnapshotsCid is a free data retrieval call binding the contract method 0xc2b97d4c.
//
// Solidity: function maxSnapshotsCid(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiSession) MaxSnapshotsCid(arg0 string, arg1 *big.Int) (string, error) {
	return _ContractApi.Contract.MaxSnapshotsCid(&_ContractApi.CallOpts, arg0, arg1)
}

// MaxSnapshotsCid is a free data retrieval call binding the contract method 0xc2b97d4c.
//
// Solidity: function maxSnapshotsCid(string , uint256 ) view returns(string)
func (_ContractApi *ContractApiCallerSession) MaxSnapshotsCid(arg0 string, arg1 *big.Int) (string, error) {
	return _ContractApi.Contract.MaxSnapshotsCid(&_ContractApi.CallOpts, arg0, arg1)
}

// MaxSnapshotsCount is a free data retrieval call binding the contract method 0xdbb54182.
//
// Solidity: function maxSnapshotsCount(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) MaxSnapshotsCount(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "maxSnapshotsCount", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxSnapshotsCount is a free data retrieval call binding the contract method 0xdbb54182.
//
// Solidity: function maxSnapshotsCount(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) MaxSnapshotsCount(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.MaxSnapshotsCount(&_ContractApi.CallOpts, arg0, arg1)
}

// MaxSnapshotsCount is a free data retrieval call binding the contract method 0xdbb54182.
//
// Solidity: function maxSnapshotsCount(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) MaxSnapshotsCount(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.MaxSnapshotsCount(&_ContractApi.CallOpts, arg0, arg1)
}

// MinSubmissionsForConsensus is a free data retrieval call binding the contract method 0x66752e1e.
//
// Solidity: function minSubmissionsForConsensus() view returns(uint256)
func (_ContractApi *ContractApiCaller) MinSubmissionsForConsensus(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "minSubmissionsForConsensus")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinSubmissionsForConsensus is a free data retrieval call binding the contract method 0x66752e1e.
//
// Solidity: function minSubmissionsForConsensus() view returns(uint256)
func (_ContractApi *ContractApiSession) MinSubmissionsForConsensus() (*big.Int, error) {
	return _ContractApi.Contract.MinSubmissionsForConsensus(&_ContractApi.CallOpts)
}

// MinSubmissionsForConsensus is a free data retrieval call binding the contract method 0x66752e1e.
//
// Solidity: function minSubmissionsForConsensus() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) MinSubmissionsForConsensus() (*big.Int, error) {
	return _ContractApi.Contract.MinSubmissionsForConsensus(&_ContractApi.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractApi *ContractApiCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractApi *ContractApiSession) Owner() (common.Address, error) {
	return _ContractApi.Contract.Owner(&_ContractApi.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractApi *ContractApiCallerSession) Owner() (common.Address, error) {
	return _ContractApi.Contract.Owner(&_ContractApi.CallOpts)
}

// ProjectFirstEpochEndHeight is a free data retrieval call binding the contract method 0x69e96804.
//
// Solidity: function projectFirstEpochEndHeight(string ) view returns(uint256)
func (_ContractApi *ContractApiCaller) ProjectFirstEpochEndHeight(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "projectFirstEpochEndHeight", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProjectFirstEpochEndHeight is a free data retrieval call binding the contract method 0x69e96804.
//
// Solidity: function projectFirstEpochEndHeight(string ) view returns(uint256)
func (_ContractApi *ContractApiSession) ProjectFirstEpochEndHeight(arg0 string) (*big.Int, error) {
	return _ContractApi.Contract.ProjectFirstEpochEndHeight(&_ContractApi.CallOpts, arg0)
}

// ProjectFirstEpochEndHeight is a free data retrieval call binding the contract method 0x69e96804.
//
// Solidity: function projectFirstEpochEndHeight(string ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) ProjectFirstEpochEndHeight(arg0 string) (*big.Int, error) {
	return _ContractApi.Contract.ProjectFirstEpochEndHeight(&_ContractApi.CallOpts, arg0)
}

// RecoverAddress is a free data retrieval call binding the contract method 0xc655d7aa.
//
// Solidity: function recoverAddress(bytes32 messageHash, bytes signature) pure returns(address)
func (_ContractApi *ContractApiCaller) RecoverAddress(opts *bind.CallOpts, messageHash [32]byte, signature []byte) (common.Address, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "recoverAddress", messageHash, signature)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RecoverAddress is a free data retrieval call binding the contract method 0xc655d7aa.
//
// Solidity: function recoverAddress(bytes32 messageHash, bytes signature) pure returns(address)
func (_ContractApi *ContractApiSession) RecoverAddress(messageHash [32]byte, signature []byte) (common.Address, error) {
	return _ContractApi.Contract.RecoverAddress(&_ContractApi.CallOpts, messageHash, signature)
}

// RecoverAddress is a free data retrieval call binding the contract method 0xc655d7aa.
//
// Solidity: function recoverAddress(bytes32 messageHash, bytes signature) pure returns(address)
func (_ContractApi *ContractApiCallerSession) RecoverAddress(messageHash [32]byte, signature []byte) (common.Address, error) {
	return _ContractApi.Contract.RecoverAddress(&_ContractApi.CallOpts, messageHash, signature)
}

// SnapshotStatus is a free data retrieval call binding the contract method 0x3aaf384d.
//
// Solidity: function snapshotStatus(string , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiCaller) SnapshotStatus(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "snapshotStatus", arg0, arg1)

	outstruct := new(struct {
		Finalized bool
		Timestamp *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Finalized = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Timestamp = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// SnapshotStatus is a free data retrieval call binding the contract method 0x3aaf384d.
//
// Solidity: function snapshotStatus(string , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiSession) SnapshotStatus(arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	return _ContractApi.Contract.SnapshotStatus(&_ContractApi.CallOpts, arg0, arg1)
}

// SnapshotStatus is a free data retrieval call binding the contract method 0x3aaf384d.
//
// Solidity: function snapshotStatus(string , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ContractApi *ContractApiCallerSession) SnapshotStatus(arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	return _ContractApi.Contract.SnapshotStatus(&_ContractApi.CallOpts, arg0, arg1)
}

// SnapshotSubmissionWindow is a free data retrieval call binding the contract method 0x059080f6.
//
// Solidity: function snapshotSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiCaller) SnapshotSubmissionWindow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "snapshotSubmissionWindow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SnapshotSubmissionWindow is a free data retrieval call binding the contract method 0x059080f6.
//
// Solidity: function snapshotSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiSession) SnapshotSubmissionWindow() (*big.Int, error) {
	return _ContractApi.Contract.SnapshotSubmissionWindow(&_ContractApi.CallOpts)
}

// SnapshotSubmissionWindow is a free data retrieval call binding the contract method 0x059080f6.
//
// Solidity: function snapshotSubmissionWindow() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SnapshotSubmissionWindow() (*big.Int, error) {
	return _ContractApi.Contract.SnapshotSubmissionWindow(&_ContractApi.CallOpts)
}

// SnapshotsReceived is a free data retrieval call binding the contract method 0x04dbb717.
//
// Solidity: function snapshotsReceived(string , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiCaller) SnapshotsReceived(opts *bind.CallOpts, arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "snapshotsReceived", arg0, arg1, arg2)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SnapshotsReceived is a free data retrieval call binding the contract method 0x04dbb717.
//
// Solidity: function snapshotsReceived(string , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiSession) SnapshotsReceived(arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	return _ContractApi.Contract.SnapshotsReceived(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceived is a free data retrieval call binding the contract method 0x04dbb717.
//
// Solidity: function snapshotsReceived(string , uint256 , address ) view returns(bool)
func (_ContractApi *ContractApiCallerSession) SnapshotsReceived(arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	return _ContractApi.Contract.SnapshotsReceived(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceivedCount is a free data retrieval call binding the contract method 0x8e07ba42.
//
// Solidity: function snapshotsReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ContractApi *ContractApiCaller) SnapshotsReceivedCount(opts *bind.CallOpts, arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "snapshotsReceivedCount", arg0, arg1, arg2)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SnapshotsReceivedCount is a free data retrieval call binding the contract method 0x8e07ba42.
//
// Solidity: function snapshotsReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ContractApi *ContractApiSession) SnapshotsReceivedCount(arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	return _ContractApi.Contract.SnapshotsReceivedCount(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceivedCount is a free data retrieval call binding the contract method 0x8e07ba42.
//
// Solidity: function snapshotsReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SnapshotsReceivedCount(arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	return _ContractApi.Contract.SnapshotsReceivedCount(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// Snapshotters is a free data retrieval call binding the contract method 0xe26ddaf3.
//
// Solidity: function snapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiCaller) Snapshotters(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "snapshotters", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Snapshotters is a free data retrieval call binding the contract method 0xe26ddaf3.
//
// Solidity: function snapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiSession) Snapshotters(arg0 common.Address) (bool, error) {
	return _ContractApi.Contract.Snapshotters(&_ContractApi.CallOpts, arg0)
}

// Snapshotters is a free data retrieval call binding the contract method 0xe26ddaf3.
//
// Solidity: function snapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiCallerSession) Snapshotters(arg0 common.Address) (bool, error) {
	return _ContractApi.Contract.Snapshotters(&_ContractApi.CallOpts, arg0)
}

// TotalSnapshotterCount is a free data retrieval call binding the contract method 0xed82f702.
//
// Solidity: function totalSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiCaller) TotalSnapshotterCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "totalSnapshotterCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSnapshotterCount is a free data retrieval call binding the contract method 0xed82f702.
//
// Solidity: function totalSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiSession) TotalSnapshotterCount() (*big.Int, error) {
	return _ContractApi.Contract.TotalSnapshotterCount(&_ContractApi.CallOpts)
}

// TotalSnapshotterCount is a free data retrieval call binding the contract method 0xed82f702.
//
// Solidity: function totalSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) TotalSnapshotterCount() (*big.Int, error) {
	return _ContractApi.Contract.TotalSnapshotterCount(&_ContractApi.CallOpts)
}

// Verify is a free data retrieval call binding the contract method 0x491612c6.
//
// Solidity: function verify((uint256) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiCaller) Verify(opts *bind.CallOpts, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte, signer common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "verify", request, signature, signer)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Verify is a free data retrieval call binding the contract method 0x491612c6.
//
// Solidity: function verify((uint256) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiSession) Verify(request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte, signer common.Address) (bool, error) {
	return _ContractApi.Contract.Verify(&_ContractApi.CallOpts, request, signature, signer)
}

// Verify is a free data retrieval call binding the contract method 0x491612c6.
//
// Solidity: function verify((uint256) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiCallerSession) Verify(request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte, signer common.Address) (bool, error) {
	return _ContractApi.Contract.Verify(&_ContractApi.CallOpts, request, signature, signer)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(address operatorAddr) returns()
func (_ContractApi *ContractApiTransactor) AddOperator(opts *bind.TransactOpts, operatorAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "addOperator", operatorAddr)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(address operatorAddr) returns()
func (_ContractApi *ContractApiSession) AddOperator(operatorAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.Contract.AddOperator(&_ContractApi.TransactOpts, operatorAddr)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(address operatorAddr) returns()
func (_ContractApi *ContractApiTransactorSession) AddOperator(operatorAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.Contract.AddOperator(&_ContractApi.TransactOpts, operatorAddr)
}

// AllowSnapshotter is a paid mutator transaction binding the contract method 0xaf3fe97f.
//
// Solidity: function allowSnapshotter(address snapshotterAddr) returns()
func (_ContractApi *ContractApiTransactor) AllowSnapshotter(opts *bind.TransactOpts, snapshotterAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "allowSnapshotter", snapshotterAddr)
}

// AllowSnapshotter is a paid mutator transaction binding the contract method 0xaf3fe97f.
//
// Solidity: function allowSnapshotter(address snapshotterAddr) returns()
func (_ContractApi *ContractApiSession) AllowSnapshotter(snapshotterAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.Contract.AllowSnapshotter(&_ContractApi.TransactOpts, snapshotterAddr)
}

// AllowSnapshotter is a paid mutator transaction binding the contract method 0xaf3fe97f.
//
// Solidity: function allowSnapshotter(address snapshotterAddr) returns()
func (_ContractApi *ContractApiTransactorSession) AllowSnapshotter(snapshotterAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.Contract.AllowSnapshotter(&_ContractApi.TransactOpts, snapshotterAddr)
}

// CommitFinalizedDAGcid is a paid mutator transaction binding the contract method 0xac595f97.
//
// Solidity: function commitFinalizedDAGcid(string projectId, uint256 dagBlockHeight, string dagCid, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactor) CommitFinalizedDAGcid(opts *bind.TransactOpts, projectId string, dagBlockHeight *big.Int, dagCid string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "commitFinalizedDAGcid", projectId, dagBlockHeight, dagCid, request, signature)
}

// CommitFinalizedDAGcid is a paid mutator transaction binding the contract method 0xac595f97.
//
// Solidity: function commitFinalizedDAGcid(string projectId, uint256 dagBlockHeight, string dagCid, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiSession) CommitFinalizedDAGcid(projectId string, dagBlockHeight *big.Int, dagCid string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.CommitFinalizedDAGcid(&_ContractApi.TransactOpts, projectId, dagBlockHeight, dagCid, request, signature)
}

// CommitFinalizedDAGcid is a paid mutator transaction binding the contract method 0xac595f97.
//
// Solidity: function commitFinalizedDAGcid(string projectId, uint256 dagBlockHeight, string dagCid, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactorSession) CommitFinalizedDAGcid(projectId string, dagBlockHeight *big.Int, dagCid string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.CommitFinalizedDAGcid(&_ContractApi.TransactOpts, projectId, dagBlockHeight, dagCid, request, signature)
}

// ForceCompleteConsensusAggregate is a paid mutator transaction binding the contract method 0xb813e704.
//
// Solidity: function forceCompleteConsensusAggregate(string projectId, uint256 DAGBlockHeight) returns()
func (_ContractApi *ContractApiTransactor) ForceCompleteConsensusAggregate(opts *bind.TransactOpts, projectId string, DAGBlockHeight *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "forceCompleteConsensusAggregate", projectId, DAGBlockHeight)
}

// ForceCompleteConsensusAggregate is a paid mutator transaction binding the contract method 0xb813e704.
//
// Solidity: function forceCompleteConsensusAggregate(string projectId, uint256 DAGBlockHeight) returns()
func (_ContractApi *ContractApiSession) ForceCompleteConsensusAggregate(projectId string, DAGBlockHeight *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceCompleteConsensusAggregate(&_ContractApi.TransactOpts, projectId, DAGBlockHeight)
}

// ForceCompleteConsensusAggregate is a paid mutator transaction binding the contract method 0xb813e704.
//
// Solidity: function forceCompleteConsensusAggregate(string projectId, uint256 DAGBlockHeight) returns()
func (_ContractApi *ContractApiTransactorSession) ForceCompleteConsensusAggregate(projectId string, DAGBlockHeight *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceCompleteConsensusAggregate(&_ContractApi.TransactOpts, projectId, DAGBlockHeight)
}

// ForceCompleteConsensusIndex is a paid mutator transaction binding the contract method 0xe6f5b00e.
//
// Solidity: function forceCompleteConsensusIndex(string projectId, uint256 DAGBlockHeight, bytes32 indexIdentifierHash) returns()
func (_ContractApi *ContractApiTransactor) ForceCompleteConsensusIndex(opts *bind.TransactOpts, projectId string, DAGBlockHeight *big.Int, indexIdentifierHash [32]byte) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "forceCompleteConsensusIndex", projectId, DAGBlockHeight, indexIdentifierHash)
}

// ForceCompleteConsensusIndex is a paid mutator transaction binding the contract method 0xe6f5b00e.
//
// Solidity: function forceCompleteConsensusIndex(string projectId, uint256 DAGBlockHeight, bytes32 indexIdentifierHash) returns()
func (_ContractApi *ContractApiSession) ForceCompleteConsensusIndex(projectId string, DAGBlockHeight *big.Int, indexIdentifierHash [32]byte) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceCompleteConsensusIndex(&_ContractApi.TransactOpts, projectId, DAGBlockHeight, indexIdentifierHash)
}

// ForceCompleteConsensusIndex is a paid mutator transaction binding the contract method 0xe6f5b00e.
//
// Solidity: function forceCompleteConsensusIndex(string projectId, uint256 DAGBlockHeight, bytes32 indexIdentifierHash) returns()
func (_ContractApi *ContractApiTransactorSession) ForceCompleteConsensusIndex(projectId string, DAGBlockHeight *big.Int, indexIdentifierHash [32]byte) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceCompleteConsensusIndex(&_ContractApi.TransactOpts, projectId, DAGBlockHeight, indexIdentifierHash)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochEnd) returns()
func (_ContractApi *ContractApiTransactor) ForceCompleteConsensusSnapshot(opts *bind.TransactOpts, projectId string, epochEnd *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "forceCompleteConsensusSnapshot", projectId, epochEnd)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochEnd) returns()
func (_ContractApi *ContractApiSession) ForceCompleteConsensusSnapshot(projectId string, epochEnd *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceCompleteConsensusSnapshot(&_ContractApi.TransactOpts, projectId, epochEnd)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochEnd) returns()
func (_ContractApi *ContractApiTransactorSession) ForceCompleteConsensusSnapshot(projectId string, epochEnd *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceCompleteConsensusSnapshot(&_ContractApi.TransactOpts, projectId, epochEnd)
}

// ReleaseEpoch is a paid mutator transaction binding the contract method 0x132c290f.
//
// Solidity: function releaseEpoch(uint256 begin, uint256 end) returns()
func (_ContractApi *ContractApiTransactor) ReleaseEpoch(opts *bind.TransactOpts, begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "releaseEpoch", begin, end)
}

// ReleaseEpoch is a paid mutator transaction binding the contract method 0x132c290f.
//
// Solidity: function releaseEpoch(uint256 begin, uint256 end) returns()
func (_ContractApi *ContractApiSession) ReleaseEpoch(begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ReleaseEpoch(&_ContractApi.TransactOpts, begin, end)
}

// ReleaseEpoch is a paid mutator transaction binding the contract method 0x132c290f.
//
// Solidity: function releaseEpoch(uint256 begin, uint256 end) returns()
func (_ContractApi *ContractApiTransactorSession) ReleaseEpoch(begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ReleaseEpoch(&_ContractApi.TransactOpts, begin, end)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(address operatorAddr) returns()
func (_ContractApi *ContractApiTransactor) RemoveOperator(opts *bind.TransactOpts, operatorAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "removeOperator", operatorAddr)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(address operatorAddr) returns()
func (_ContractApi *ContractApiSession) RemoveOperator(operatorAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.Contract.RemoveOperator(&_ContractApi.TransactOpts, operatorAddr)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(address operatorAddr) returns()
func (_ContractApi *ContractApiTransactorSession) RemoveOperator(operatorAddr common.Address) (*types.Transaction, error) {
	return _ContractApi.Contract.RemoveOperator(&_ContractApi.TransactOpts, operatorAddr)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractApi *ContractApiTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractApi *ContractApiSession) RenounceOwnership() (*types.Transaction, error) {
	return _ContractApi.Contract.RenounceOwnership(&_ContractApi.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractApi *ContractApiTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ContractApi.Contract.RenounceOwnership(&_ContractApi.TransactOpts)
}

// SubmitAggregate is a paid mutator transaction binding the contract method 0x47a1c598.
//
// Solidity: function submitAggregate(string snapshotCid, uint256 DAGBlockHeight, string projectId, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactor) SubmitAggregate(opts *bind.TransactOpts, snapshotCid string, DAGBlockHeight *big.Int, projectId string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "submitAggregate", snapshotCid, DAGBlockHeight, projectId, request, signature)
}

// SubmitAggregate is a paid mutator transaction binding the contract method 0x47a1c598.
//
// Solidity: function submitAggregate(string snapshotCid, uint256 DAGBlockHeight, string projectId, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiSession) SubmitAggregate(snapshotCid string, DAGBlockHeight *big.Int, projectId string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitAggregate(&_ContractApi.TransactOpts, snapshotCid, DAGBlockHeight, projectId, request, signature)
}

// SubmitAggregate is a paid mutator transaction binding the contract method 0x47a1c598.
//
// Solidity: function submitAggregate(string snapshotCid, uint256 DAGBlockHeight, string projectId, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactorSession) SubmitAggregate(snapshotCid string, DAGBlockHeight *big.Int, projectId string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitAggregate(&_ContractApi.TransactOpts, snapshotCid, DAGBlockHeight, projectId, request, signature)
}

// SubmitIndex is a paid mutator transaction binding the contract method 0x3c473033.
//
// Solidity: function submitIndex(string projectId, uint256 DAGBlockHeight, uint256 indexTailDAGBlockHeight, uint256 tailBlockEpochSourceChainHeight, bytes32 indexIdentifierHash, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactor) SubmitIndex(opts *bind.TransactOpts, projectId string, DAGBlockHeight *big.Int, indexTailDAGBlockHeight *big.Int, tailBlockEpochSourceChainHeight *big.Int, indexIdentifierHash [32]byte, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "submitIndex", projectId, DAGBlockHeight, indexTailDAGBlockHeight, tailBlockEpochSourceChainHeight, indexIdentifierHash, request, signature)
}

// SubmitIndex is a paid mutator transaction binding the contract method 0x3c473033.
//
// Solidity: function submitIndex(string projectId, uint256 DAGBlockHeight, uint256 indexTailDAGBlockHeight, uint256 tailBlockEpochSourceChainHeight, bytes32 indexIdentifierHash, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiSession) SubmitIndex(projectId string, DAGBlockHeight *big.Int, indexTailDAGBlockHeight *big.Int, tailBlockEpochSourceChainHeight *big.Int, indexIdentifierHash [32]byte, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitIndex(&_ContractApi.TransactOpts, projectId, DAGBlockHeight, indexTailDAGBlockHeight, tailBlockEpochSourceChainHeight, indexIdentifierHash, request, signature)
}

// SubmitIndex is a paid mutator transaction binding the contract method 0x3c473033.
//
// Solidity: function submitIndex(string projectId, uint256 DAGBlockHeight, uint256 indexTailDAGBlockHeight, uint256 tailBlockEpochSourceChainHeight, bytes32 indexIdentifierHash, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactorSession) SubmitIndex(projectId string, DAGBlockHeight *big.Int, indexTailDAGBlockHeight *big.Int, tailBlockEpochSourceChainHeight *big.Int, indexIdentifierHash [32]byte, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitIndex(&_ContractApi.TransactOpts, projectId, DAGBlockHeight, indexTailDAGBlockHeight, tailBlockEpochSourceChainHeight, indexIdentifierHash, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0xbef76fb1.
//
// Solidity: function submitSnapshot(string snapshotCid, uint256 epochEnd, string projectId, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactor) SubmitSnapshot(opts *bind.TransactOpts, snapshotCid string, epochEnd *big.Int, projectId string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "submitSnapshot", snapshotCid, epochEnd, projectId, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0xbef76fb1.
//
// Solidity: function submitSnapshot(string snapshotCid, uint256 epochEnd, string projectId, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiSession) SubmitSnapshot(snapshotCid string, epochEnd *big.Int, projectId string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitSnapshot(&_ContractApi.TransactOpts, snapshotCid, epochEnd, projectId, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0xbef76fb1.
//
// Solidity: function submitSnapshot(string snapshotCid, uint256 epochEnd, string projectId, (uint256) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactorSession) SubmitSnapshot(snapshotCid string, epochEnd *big.Int, projectId string, request AuditRecordStoreDynamicSnapshottersWithIndexingRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitSnapshot(&_ContractApi.TransactOpts, snapshotCid, epochEnd, projectId, request, signature)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractApi *ContractApiTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractApi *ContractApiSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ContractApi.Contract.TransferOwnership(&_ContractApi.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractApi *ContractApiTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ContractApi.Contract.TransferOwnership(&_ContractApi.TransactOpts, newOwner)
}

// UpdateAggregateSubmissionWindow is a paid mutator transaction binding the contract method 0x544c5057.
//
// Solidity: function updateAggregateSubmissionWindow(uint256 newAggregateSubmissionWindow) returns()
func (_ContractApi *ContractApiTransactor) UpdateAggregateSubmissionWindow(opts *bind.TransactOpts, newAggregateSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateAggregateSubmissionWindow", newAggregateSubmissionWindow)
}

// UpdateAggregateSubmissionWindow is a paid mutator transaction binding the contract method 0x544c5057.
//
// Solidity: function updateAggregateSubmissionWindow(uint256 newAggregateSubmissionWindow) returns()
func (_ContractApi *ContractApiSession) UpdateAggregateSubmissionWindow(newAggregateSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateAggregateSubmissionWindow(&_ContractApi.TransactOpts, newAggregateSubmissionWindow)
}

// UpdateAggregateSubmissionWindow is a paid mutator transaction binding the contract method 0x544c5057.
//
// Solidity: function updateAggregateSubmissionWindow(uint256 newAggregateSubmissionWindow) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateAggregateSubmissionWindow(newAggregateSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateAggregateSubmissionWindow(&_ContractApi.TransactOpts, newAggregateSubmissionWindow)
}

// UpdateIndexSubmissionWindow is a paid mutator transaction binding the contract method 0x50e15e55.
//
// Solidity: function updateIndexSubmissionWindow(uint256 newIndexSubmissionWindow) returns()
func (_ContractApi *ContractApiTransactor) UpdateIndexSubmissionWindow(opts *bind.TransactOpts, newIndexSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateIndexSubmissionWindow", newIndexSubmissionWindow)
}

// UpdateIndexSubmissionWindow is a paid mutator transaction binding the contract method 0x50e15e55.
//
// Solidity: function updateIndexSubmissionWindow(uint256 newIndexSubmissionWindow) returns()
func (_ContractApi *ContractApiSession) UpdateIndexSubmissionWindow(newIndexSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateIndexSubmissionWindow(&_ContractApi.TransactOpts, newIndexSubmissionWindow)
}

// UpdateIndexSubmissionWindow is a paid mutator transaction binding the contract method 0x50e15e55.
//
// Solidity: function updateIndexSubmissionWindow(uint256 newIndexSubmissionWindow) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateIndexSubmissionWindow(newIndexSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateIndexSubmissionWindow(&_ContractApi.TransactOpts, newIndexSubmissionWindow)
}

// UpdateMinSnapshottersForConsensus is a paid mutator transaction binding the contract method 0x38deacd3.
//
// Solidity: function updateMinSnapshottersForConsensus(uint256 _minSubmissionsForConsensus) returns()
func (_ContractApi *ContractApiTransactor) UpdateMinSnapshottersForConsensus(opts *bind.TransactOpts, _minSubmissionsForConsensus *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateMinSnapshottersForConsensus", _minSubmissionsForConsensus)
}

// UpdateMinSnapshottersForConsensus is a paid mutator transaction binding the contract method 0x38deacd3.
//
// Solidity: function updateMinSnapshottersForConsensus(uint256 _minSubmissionsForConsensus) returns()
func (_ContractApi *ContractApiSession) UpdateMinSnapshottersForConsensus(_minSubmissionsForConsensus *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateMinSnapshottersForConsensus(&_ContractApi.TransactOpts, _minSubmissionsForConsensus)
}

// UpdateMinSnapshottersForConsensus is a paid mutator transaction binding the contract method 0x38deacd3.
//
// Solidity: function updateMinSnapshottersForConsensus(uint256 _minSubmissionsForConsensus) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateMinSnapshottersForConsensus(_minSubmissionsForConsensus *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateMinSnapshottersForConsensus(&_ContractApi.TransactOpts, _minSubmissionsForConsensus)
}

// UpdateSnapshotSubmissionWindow is a paid mutator transaction binding the contract method 0x9b2f89ce.
//
// Solidity: function updateSnapshotSubmissionWindow(uint256 newsnapshotSubmissionWindow) returns()
func (_ContractApi *ContractApiTransactor) UpdateSnapshotSubmissionWindow(opts *bind.TransactOpts, newsnapshotSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateSnapshotSubmissionWindow", newsnapshotSubmissionWindow)
}

// UpdateSnapshotSubmissionWindow is a paid mutator transaction binding the contract method 0x9b2f89ce.
//
// Solidity: function updateSnapshotSubmissionWindow(uint256 newsnapshotSubmissionWindow) returns()
func (_ContractApi *ContractApiSession) UpdateSnapshotSubmissionWindow(newsnapshotSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateSnapshotSubmissionWindow(&_ContractApi.TransactOpts, newsnapshotSubmissionWindow)
}

// UpdateSnapshotSubmissionWindow is a paid mutator transaction binding the contract method 0x9b2f89ce.
//
// Solidity: function updateSnapshotSubmissionWindow(uint256 newsnapshotSubmissionWindow) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateSnapshotSubmissionWindow(newsnapshotSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateSnapshotSubmissionWindow(&_ContractApi.TransactOpts, newsnapshotSubmissionWindow)
}

// ContractApiAdminModuleUpdatedIterator is returned from FilterAdminModuleUpdated and is used to iterate over the raw logs and unpacked data for AdminModuleUpdated events raised by the ContractApi contract.
type ContractApiAdminModuleUpdatedIterator struct {
	Event *ContractApiAdminModuleUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiAdminModuleUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiAdminModuleUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiAdminModuleUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiAdminModuleUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiAdminModuleUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiAdminModuleUpdated represents a AdminModuleUpdated event raised by the ContractApi contract.
type ContractApiAdminModuleUpdated struct {
	AdminModuleAddress common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterAdminModuleUpdated is a free log retrieval operation binding the contract event 0xa707f008b0fef2b9c7bb05b46753c01a64f4ed9948f7a22958160c347baeefcf.
//
// Solidity: event AdminModuleUpdated(address adminModuleAddress)
func (_ContractApi *ContractApiFilterer) FilterAdminModuleUpdated(opts *bind.FilterOpts) (*ContractApiAdminModuleUpdatedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "AdminModuleUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractApiAdminModuleUpdatedIterator{contract: _ContractApi.contract, event: "AdminModuleUpdated", logs: logs, sub: sub}, nil
}

// WatchAdminModuleUpdated is a free log subscription operation binding the contract event 0xa707f008b0fef2b9c7bb05b46753c01a64f4ed9948f7a22958160c347baeefcf.
//
// Solidity: event AdminModuleUpdated(address adminModuleAddress)
func (_ContractApi *ContractApiFilterer) WatchAdminModuleUpdated(opts *bind.WatchOpts, sink chan<- *ContractApiAdminModuleUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "AdminModuleUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiAdminModuleUpdated)
				if err := _ContractApi.contract.UnpackLog(event, "AdminModuleUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAdminModuleUpdated is a log parse operation binding the contract event 0xa707f008b0fef2b9c7bb05b46753c01a64f4ed9948f7a22958160c347baeefcf.
//
// Solidity: event AdminModuleUpdated(address adminModuleAddress)
func (_ContractApi *ContractApiFilterer) ParseAdminModuleUpdated(log types.Log) (*ContractApiAdminModuleUpdated, error) {
	event := new(ContractApiAdminModuleUpdated)
	if err := _ContractApi.contract.UnpackLog(event, "AdminModuleUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiAggregateDagCidFinalizedIterator is returned from FilterAggregateDagCidFinalized and is used to iterate over the raw logs and unpacked data for AggregateDagCidFinalized events raised by the ContractApi contract.
type ContractApiAggregateDagCidFinalizedIterator struct {
	Event *ContractApiAggregateDagCidFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiAggregateDagCidFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiAggregateDagCidFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiAggregateDagCidFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiAggregateDagCidFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiAggregateDagCidFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiAggregateDagCidFinalized represents a AggregateDagCidFinalized event raised by the ContractApi contract.
type ContractApiAggregateDagCidFinalized struct {
	ProjectId      string
	DAGBlockHeight *big.Int
	DagCid         string
	Timestamp      *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterAggregateDagCidFinalized is a free log retrieval operation binding the contract event 0x7e08f287b45a9a29fc7082b55216e4e1e9c690318829f3c0d55e62b090f9c941.
//
// Solidity: event AggregateDagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterAggregateDagCidFinalized(opts *bind.FilterOpts, DAGBlockHeight []*big.Int, timestamp []*big.Int) (*ContractApiAggregateDagCidFinalizedIterator, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "AggregateDagCidFinalized", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiAggregateDagCidFinalizedIterator{contract: _ContractApi.contract, event: "AggregateDagCidFinalized", logs: logs, sub: sub}, nil
}

// WatchAggregateDagCidFinalized is a free log subscription operation binding the contract event 0x7e08f287b45a9a29fc7082b55216e4e1e9c690318829f3c0d55e62b090f9c941.
//
// Solidity: event AggregateDagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchAggregateDagCidFinalized(opts *bind.WatchOpts, sink chan<- *ContractApiAggregateDagCidFinalized, DAGBlockHeight []*big.Int, timestamp []*big.Int) (event.Subscription, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "AggregateDagCidFinalized", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiAggregateDagCidFinalized)
				if err := _ContractApi.contract.UnpackLog(event, "AggregateDagCidFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAggregateDagCidFinalized is a log parse operation binding the contract event 0x7e08f287b45a9a29fc7082b55216e4e1e9c690318829f3c0d55e62b090f9c941.
//
// Solidity: event AggregateDagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseAggregateDagCidFinalized(log types.Log) (*ContractApiAggregateDagCidFinalized, error) {
	event := new(ContractApiAggregateDagCidFinalized)
	if err := _ContractApi.contract.UnpackLog(event, "AggregateDagCidFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiAggregateDagCidSubmittedIterator is returned from FilterAggregateDagCidSubmitted and is used to iterate over the raw logs and unpacked data for AggregateDagCidSubmitted events raised by the ContractApi contract.
type ContractApiAggregateDagCidSubmittedIterator struct {
	Event *ContractApiAggregateDagCidSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiAggregateDagCidSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiAggregateDagCidSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiAggregateDagCidSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiAggregateDagCidSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiAggregateDagCidSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiAggregateDagCidSubmitted represents a AggregateDagCidSubmitted event raised by the ContractApi contract.
type ContractApiAggregateDagCidSubmitted struct {
	ProjectId      string
	DAGBlockHeight *big.Int
	DagCid         string
	Timestamp      *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterAggregateDagCidSubmitted is a free log retrieval operation binding the contract event 0x2f1f55bc0f8bc1a85d65c0e91f1377ee9ce42132d2ca68fe779800066f45faef.
//
// Solidity: event AggregateDagCidSubmitted(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterAggregateDagCidSubmitted(opts *bind.FilterOpts, DAGBlockHeight []*big.Int, timestamp []*big.Int) (*ContractApiAggregateDagCidSubmittedIterator, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "AggregateDagCidSubmitted", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiAggregateDagCidSubmittedIterator{contract: _ContractApi.contract, event: "AggregateDagCidSubmitted", logs: logs, sub: sub}, nil
}

// WatchAggregateDagCidSubmitted is a free log subscription operation binding the contract event 0x2f1f55bc0f8bc1a85d65c0e91f1377ee9ce42132d2ca68fe779800066f45faef.
//
// Solidity: event AggregateDagCidSubmitted(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchAggregateDagCidSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiAggregateDagCidSubmitted, DAGBlockHeight []*big.Int, timestamp []*big.Int) (event.Subscription, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "AggregateDagCidSubmitted", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiAggregateDagCidSubmitted)
				if err := _ContractApi.contract.UnpackLog(event, "AggregateDagCidSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAggregateDagCidSubmitted is a log parse operation binding the contract event 0x2f1f55bc0f8bc1a85d65c0e91f1377ee9ce42132d2ca68fe779800066f45faef.
//
// Solidity: event AggregateDagCidSubmitted(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseAggregateDagCidSubmitted(log types.Log) (*ContractApiAggregateDagCidSubmitted, error) {
	event := new(ContractApiAggregateDagCidSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "AggregateDagCidSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiAggregateFinalizedIterator is returned from FilterAggregateFinalized and is used to iterate over the raw logs and unpacked data for AggregateFinalized events raised by the ContractApi contract.
type ContractApiAggregateFinalizedIterator struct {
	Event *ContractApiAggregateFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiAggregateFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiAggregateFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiAggregateFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiAggregateFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiAggregateFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiAggregateFinalized represents a AggregateFinalized event raised by the ContractApi contract.
type ContractApiAggregateFinalized struct {
	EpochEnd     *big.Int
	ProjectId    string
	AggregateCid string
	Timestamp    *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterAggregateFinalized is a free log retrieval operation binding the contract event 0xb71d635991175a1ba2ee45a5f9d04619e3fe569effc907b23c9e3069f491fd5e.
//
// Solidity: event AggregateFinalized(uint256 epochEnd, string projectId, string aggregateCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterAggregateFinalized(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiAggregateFinalizedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "AggregateFinalized", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiAggregateFinalizedIterator{contract: _ContractApi.contract, event: "AggregateFinalized", logs: logs, sub: sub}, nil
}

// WatchAggregateFinalized is a free log subscription operation binding the contract event 0xb71d635991175a1ba2ee45a5f9d04619e3fe569effc907b23c9e3069f491fd5e.
//
// Solidity: event AggregateFinalized(uint256 epochEnd, string projectId, string aggregateCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchAggregateFinalized(opts *bind.WatchOpts, sink chan<- *ContractApiAggregateFinalized, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "AggregateFinalized", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiAggregateFinalized)
				if err := _ContractApi.contract.UnpackLog(event, "AggregateFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAggregateFinalized is a log parse operation binding the contract event 0xb71d635991175a1ba2ee45a5f9d04619e3fe569effc907b23c9e3069f491fd5e.
//
// Solidity: event AggregateFinalized(uint256 epochEnd, string projectId, string aggregateCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseAggregateFinalized(log types.Log) (*ContractApiAggregateFinalized, error) {
	event := new(ContractApiAggregateFinalized)
	if err := _ContractApi.contract.UnpackLog(event, "AggregateFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiAggregateSubmittedIterator is returned from FilterAggregateSubmitted and is used to iterate over the raw logs and unpacked data for AggregateSubmitted events raised by the ContractApi contract.
type ContractApiAggregateSubmittedIterator struct {
	Event *ContractApiAggregateSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiAggregateSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiAggregateSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiAggregateSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiAggregateSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiAggregateSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiAggregateSubmitted represents a AggregateSubmitted event raised by the ContractApi contract.
type ContractApiAggregateSubmitted struct {
	SnapshotterAddr common.Address
	AggregateCid    string
	EpochEnd        *big.Int
	ProjectId       string
	Timestamp       *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterAggregateSubmitted is a free log retrieval operation binding the contract event 0x777f7e9e6ebfa1545794ad3b649db5f1dca1edb9e7fd0992623f2c487b8ef41c.
//
// Solidity: event AggregateSubmitted(address snapshotterAddr, string aggregateCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterAggregateSubmitted(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiAggregateSubmittedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "AggregateSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiAggregateSubmittedIterator{contract: _ContractApi.contract, event: "AggregateSubmitted", logs: logs, sub: sub}, nil
}

// WatchAggregateSubmitted is a free log subscription operation binding the contract event 0x777f7e9e6ebfa1545794ad3b649db5f1dca1edb9e7fd0992623f2c487b8ef41c.
//
// Solidity: event AggregateSubmitted(address snapshotterAddr, string aggregateCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchAggregateSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiAggregateSubmitted, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "AggregateSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiAggregateSubmitted)
				if err := _ContractApi.contract.UnpackLog(event, "AggregateSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAggregateSubmitted is a log parse operation binding the contract event 0x777f7e9e6ebfa1545794ad3b649db5f1dca1edb9e7fd0992623f2c487b8ef41c.
//
// Solidity: event AggregateSubmitted(address snapshotterAddr, string aggregateCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseAggregateSubmitted(log types.Log) (*ContractApiAggregateSubmitted, error) {
	event := new(ContractApiAggregateSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "AggregateSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiDagCidFinalizedIterator is returned from FilterDagCidFinalized and is used to iterate over the raw logs and unpacked data for DagCidFinalized events raised by the ContractApi contract.
type ContractApiDagCidFinalizedIterator struct {
	Event *ContractApiDagCidFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiDagCidFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiDagCidFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiDagCidFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiDagCidFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiDagCidFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiDagCidFinalized represents a DagCidFinalized event raised by the ContractApi contract.
type ContractApiDagCidFinalized struct {
	ProjectId      string
	DAGBlockHeight *big.Int
	DagCid         string
	Timestamp      *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterDagCidFinalized is a free log retrieval operation binding the contract event 0x8cf628c80bd11df0951e0e4a59a2e0980783df0287f45147601403c13ae4f1a3.
//
// Solidity: event DagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterDagCidFinalized(opts *bind.FilterOpts, DAGBlockHeight []*big.Int, timestamp []*big.Int) (*ContractApiDagCidFinalizedIterator, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "DagCidFinalized", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiDagCidFinalizedIterator{contract: _ContractApi.contract, event: "DagCidFinalized", logs: logs, sub: sub}, nil
}

// WatchDagCidFinalized is a free log subscription operation binding the contract event 0x8cf628c80bd11df0951e0e4a59a2e0980783df0287f45147601403c13ae4f1a3.
//
// Solidity: event DagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchDagCidFinalized(opts *bind.WatchOpts, sink chan<- *ContractApiDagCidFinalized, DAGBlockHeight []*big.Int, timestamp []*big.Int) (event.Subscription, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "DagCidFinalized", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiDagCidFinalized)
				if err := _ContractApi.contract.UnpackLog(event, "DagCidFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDagCidFinalized is a log parse operation binding the contract event 0x8cf628c80bd11df0951e0e4a59a2e0980783df0287f45147601403c13ae4f1a3.
//
// Solidity: event DagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseDagCidFinalized(log types.Log) (*ContractApiDagCidFinalized, error) {
	event := new(ContractApiDagCidFinalized)
	if err := _ContractApi.contract.UnpackLog(event, "DagCidFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiDelayedAggregateSubmittedIterator is returned from FilterDelayedAggregateSubmitted and is used to iterate over the raw logs and unpacked data for DelayedAggregateSubmitted events raised by the ContractApi contract.
type ContractApiDelayedAggregateSubmittedIterator struct {
	Event *ContractApiDelayedAggregateSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiDelayedAggregateSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiDelayedAggregateSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiDelayedAggregateSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiDelayedAggregateSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiDelayedAggregateSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiDelayedAggregateSubmitted represents a DelayedAggregateSubmitted event raised by the ContractApi contract.
type ContractApiDelayedAggregateSubmitted struct {
	SnapshotterAddr common.Address
	AggregateCid    string
	EpochEnd        *big.Int
	ProjectId       string
	Timestamp       *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterDelayedAggregateSubmitted is a free log retrieval operation binding the contract event 0x3543cb13506e7dcda3739006e8a844427cf59c786cd04268c2875ffc0d1a8848.
//
// Solidity: event DelayedAggregateSubmitted(address snapshotterAddr, string aggregateCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterDelayedAggregateSubmitted(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiDelayedAggregateSubmittedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "DelayedAggregateSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiDelayedAggregateSubmittedIterator{contract: _ContractApi.contract, event: "DelayedAggregateSubmitted", logs: logs, sub: sub}, nil
}

// WatchDelayedAggregateSubmitted is a free log subscription operation binding the contract event 0x3543cb13506e7dcda3739006e8a844427cf59c786cd04268c2875ffc0d1a8848.
//
// Solidity: event DelayedAggregateSubmitted(address snapshotterAddr, string aggregateCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchDelayedAggregateSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiDelayedAggregateSubmitted, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "DelayedAggregateSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiDelayedAggregateSubmitted)
				if err := _ContractApi.contract.UnpackLog(event, "DelayedAggregateSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelayedAggregateSubmitted is a log parse operation binding the contract event 0x3543cb13506e7dcda3739006e8a844427cf59c786cd04268c2875ffc0d1a8848.
//
// Solidity: event DelayedAggregateSubmitted(address snapshotterAddr, string aggregateCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseDelayedAggregateSubmitted(log types.Log) (*ContractApiDelayedAggregateSubmitted, error) {
	event := new(ContractApiDelayedAggregateSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "DelayedAggregateSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiDelayedIndexSubmittedIterator is returned from FilterDelayedIndexSubmitted and is used to iterate over the raw logs and unpacked data for DelayedIndexSubmitted events raised by the ContractApi contract.
type ContractApiDelayedIndexSubmittedIterator struct {
	Event *ContractApiDelayedIndexSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiDelayedIndexSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiDelayedIndexSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiDelayedIndexSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiDelayedIndexSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiDelayedIndexSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiDelayedIndexSubmitted represents a DelayedIndexSubmitted event raised by the ContractApi contract.
type ContractApiDelayedIndexSubmitted struct {
	SnapshotterAddr         common.Address
	IndexTailDAGBlockHeight *big.Int
	DAGBlockHeight          *big.Int
	ProjectId               string
	IndexIdentifierHash     [32]byte
	Timestamp               *big.Int
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterDelayedIndexSubmitted is a free log retrieval operation binding the contract event 0x5c43e477573c41857016e568d10ab54b846a5ea332b30d079a8799425114301e.
//
// Solidity: event DelayedIndexSubmitted(address snapshotterAddr, uint256 indexTailDAGBlockHeight, uint256 DAGBlockHeight, string projectId, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterDelayedIndexSubmitted(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiDelayedIndexSubmittedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "DelayedIndexSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiDelayedIndexSubmittedIterator{contract: _ContractApi.contract, event: "DelayedIndexSubmitted", logs: logs, sub: sub}, nil
}

// WatchDelayedIndexSubmitted is a free log subscription operation binding the contract event 0x5c43e477573c41857016e568d10ab54b846a5ea332b30d079a8799425114301e.
//
// Solidity: event DelayedIndexSubmitted(address snapshotterAddr, uint256 indexTailDAGBlockHeight, uint256 DAGBlockHeight, string projectId, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchDelayedIndexSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiDelayedIndexSubmitted, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "DelayedIndexSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiDelayedIndexSubmitted)
				if err := _ContractApi.contract.UnpackLog(event, "DelayedIndexSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelayedIndexSubmitted is a log parse operation binding the contract event 0x5c43e477573c41857016e568d10ab54b846a5ea332b30d079a8799425114301e.
//
// Solidity: event DelayedIndexSubmitted(address snapshotterAddr, uint256 indexTailDAGBlockHeight, uint256 DAGBlockHeight, string projectId, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseDelayedIndexSubmitted(log types.Log) (*ContractApiDelayedIndexSubmitted, error) {
	event := new(ContractApiDelayedIndexSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "DelayedIndexSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiDelayedSnapshotSubmittedIterator is returned from FilterDelayedSnapshotSubmitted and is used to iterate over the raw logs and unpacked data for DelayedSnapshotSubmitted events raised by the ContractApi contract.
type ContractApiDelayedSnapshotSubmittedIterator struct {
	Event *ContractApiDelayedSnapshotSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiDelayedSnapshotSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiDelayedSnapshotSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiDelayedSnapshotSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiDelayedSnapshotSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiDelayedSnapshotSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiDelayedSnapshotSubmitted represents a DelayedSnapshotSubmitted event raised by the ContractApi contract.
type ContractApiDelayedSnapshotSubmitted struct {
	SnapshotterAddr common.Address
	SnapshotCid     string
	EpochEnd        *big.Int
	ProjectId       string
	Timestamp       *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterDelayedSnapshotSubmitted is a free log retrieval operation binding the contract event 0xa4b1762053ee970f50692b6936d4e58a9a01291449e4da16bdf758891c8de752.
//
// Solidity: event DelayedSnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterDelayedSnapshotSubmitted(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiDelayedSnapshotSubmittedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "DelayedSnapshotSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiDelayedSnapshotSubmittedIterator{contract: _ContractApi.contract, event: "DelayedSnapshotSubmitted", logs: logs, sub: sub}, nil
}

// WatchDelayedSnapshotSubmitted is a free log subscription operation binding the contract event 0xa4b1762053ee970f50692b6936d4e58a9a01291449e4da16bdf758891c8de752.
//
// Solidity: event DelayedSnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchDelayedSnapshotSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiDelayedSnapshotSubmitted, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "DelayedSnapshotSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiDelayedSnapshotSubmitted)
				if err := _ContractApi.contract.UnpackLog(event, "DelayedSnapshotSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelayedSnapshotSubmitted is a log parse operation binding the contract event 0xa4b1762053ee970f50692b6936d4e58a9a01291449e4da16bdf758891c8de752.
//
// Solidity: event DelayedSnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseDelayedSnapshotSubmitted(log types.Log) (*ContractApiDelayedSnapshotSubmitted, error) {
	event := new(ContractApiDelayedSnapshotSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "DelayedSnapshotSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiEpochReleasedIterator is returned from FilterEpochReleased and is used to iterate over the raw logs and unpacked data for EpochReleased events raised by the ContractApi contract.
type ContractApiEpochReleasedIterator struct {
	Event *ContractApiEpochReleased // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiEpochReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiEpochReleased)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiEpochReleased)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiEpochReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiEpochReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiEpochReleased represents a EpochReleased event raised by the ContractApi contract.
type ContractApiEpochReleased struct {
	Begin     *big.Int
	End       *big.Int
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEpochReleased is a free log retrieval operation binding the contract event 0x0fa2c0974a4f89f95d22cc151e41167436ff87e9acb951394b63235bcfc8726b.
//
// Solidity: event EpochReleased(uint256 begin, uint256 end, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterEpochReleased(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiEpochReleasedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "EpochReleased", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiEpochReleasedIterator{contract: _ContractApi.contract, event: "EpochReleased", logs: logs, sub: sub}, nil
}

// WatchEpochReleased is a free log subscription operation binding the contract event 0x0fa2c0974a4f89f95d22cc151e41167436ff87e9acb951394b63235bcfc8726b.
//
// Solidity: event EpochReleased(uint256 begin, uint256 end, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchEpochReleased(opts *bind.WatchOpts, sink chan<- *ContractApiEpochReleased, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "EpochReleased", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiEpochReleased)
				if err := _ContractApi.contract.UnpackLog(event, "EpochReleased", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEpochReleased is a log parse operation binding the contract event 0x0fa2c0974a4f89f95d22cc151e41167436ff87e9acb951394b63235bcfc8726b.
//
// Solidity: event EpochReleased(uint256 begin, uint256 end, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseEpochReleased(log types.Log) (*ContractApiEpochReleased, error) {
	event := new(ContractApiEpochReleased)
	if err := _ContractApi.contract.UnpackLog(event, "EpochReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiIndexFinalizedIterator is returned from FilterIndexFinalized and is used to iterate over the raw logs and unpacked data for IndexFinalized events raised by the ContractApi contract.
type ContractApiIndexFinalizedIterator struct {
	Event *ContractApiIndexFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiIndexFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiIndexFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiIndexFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiIndexFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiIndexFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiIndexFinalized represents a IndexFinalized event raised by the ContractApi contract.
type ContractApiIndexFinalized struct {
	ProjectId                       string
	DAGBlockHeight                  *big.Int
	IndexTailDAGBlockHeight         *big.Int
	TailBlockEpochSourceChainHeight *big.Int
	IndexIdentifierHash             [32]byte
	Timestamp                       *big.Int
	Raw                             types.Log // Blockchain specific contextual infos
}

// FilterIndexFinalized is a free log retrieval operation binding the contract event 0x17687f2ab028c4b1c6f700382e58c9a8b1dab588a77139850c1f1d6f2cf2e763.
//
// Solidity: event IndexFinalized(string projectId, uint256 DAGBlockHeight, uint256 indexTailDAGBlockHeight, uint256 tailBlockEpochSourceChainHeight, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterIndexFinalized(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiIndexFinalizedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "IndexFinalized", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiIndexFinalizedIterator{contract: _ContractApi.contract, event: "IndexFinalized", logs: logs, sub: sub}, nil
}

// WatchIndexFinalized is a free log subscription operation binding the contract event 0x17687f2ab028c4b1c6f700382e58c9a8b1dab588a77139850c1f1d6f2cf2e763.
//
// Solidity: event IndexFinalized(string projectId, uint256 DAGBlockHeight, uint256 indexTailDAGBlockHeight, uint256 tailBlockEpochSourceChainHeight, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchIndexFinalized(opts *bind.WatchOpts, sink chan<- *ContractApiIndexFinalized, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "IndexFinalized", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiIndexFinalized)
				if err := _ContractApi.contract.UnpackLog(event, "IndexFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseIndexFinalized is a log parse operation binding the contract event 0x17687f2ab028c4b1c6f700382e58c9a8b1dab588a77139850c1f1d6f2cf2e763.
//
// Solidity: event IndexFinalized(string projectId, uint256 DAGBlockHeight, uint256 indexTailDAGBlockHeight, uint256 tailBlockEpochSourceChainHeight, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseIndexFinalized(log types.Log) (*ContractApiIndexFinalized, error) {
	event := new(ContractApiIndexFinalized)
	if err := _ContractApi.contract.UnpackLog(event, "IndexFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiIndexSubmittedIterator is returned from FilterIndexSubmitted and is used to iterate over the raw logs and unpacked data for IndexSubmitted events raised by the ContractApi contract.
type ContractApiIndexSubmittedIterator struct {
	Event *ContractApiIndexSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiIndexSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiIndexSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiIndexSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiIndexSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiIndexSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiIndexSubmitted represents a IndexSubmitted event raised by the ContractApi contract.
type ContractApiIndexSubmitted struct {
	SnapshotterAddr         common.Address
	IndexTailDAGBlockHeight *big.Int
	DAGBlockHeight          *big.Int
	ProjectId               string
	IndexIdentifierHash     [32]byte
	Timestamp               *big.Int
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterIndexSubmitted is a free log retrieval operation binding the contract event 0x604843ad7cdbf0c4463153374bb70a70c9b1aa622007dbc0c74d81cfb96f4bd8.
//
// Solidity: event IndexSubmitted(address snapshotterAddr, uint256 indexTailDAGBlockHeight, uint256 DAGBlockHeight, string projectId, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterIndexSubmitted(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiIndexSubmittedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "IndexSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiIndexSubmittedIterator{contract: _ContractApi.contract, event: "IndexSubmitted", logs: logs, sub: sub}, nil
}

// WatchIndexSubmitted is a free log subscription operation binding the contract event 0x604843ad7cdbf0c4463153374bb70a70c9b1aa622007dbc0c74d81cfb96f4bd8.
//
// Solidity: event IndexSubmitted(address snapshotterAddr, uint256 indexTailDAGBlockHeight, uint256 DAGBlockHeight, string projectId, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchIndexSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiIndexSubmitted, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "IndexSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiIndexSubmitted)
				if err := _ContractApi.contract.UnpackLog(event, "IndexSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseIndexSubmitted is a log parse operation binding the contract event 0x604843ad7cdbf0c4463153374bb70a70c9b1aa622007dbc0c74d81cfb96f4bd8.
//
// Solidity: event IndexSubmitted(address snapshotterAddr, uint256 indexTailDAGBlockHeight, uint256 DAGBlockHeight, string projectId, bytes32 indexIdentifierHash, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseIndexSubmitted(log types.Log) (*ContractApiIndexSubmitted, error) {
	event := new(ContractApiIndexSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "IndexSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ContractApi contract.
type ContractApiOwnershipTransferredIterator struct {
	Event *ContractApiOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiOwnershipTransferred represents a OwnershipTransferred event raised by the ContractApi contract.
type ContractApiOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractApi *ContractApiFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractApiOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiOwnershipTransferredIterator{contract: _ContractApi.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractApi *ContractApiFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractApiOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiOwnershipTransferred)
				if err := _ContractApi.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractApi *ContractApiFilterer) ParseOwnershipTransferred(log types.Log) (*ContractApiOwnershipTransferred, error) {
	event := new(ContractApiOwnershipTransferred)
	if err := _ContractApi.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiSnapshotDagCidFinalizedIterator is returned from FilterSnapshotDagCidFinalized and is used to iterate over the raw logs and unpacked data for SnapshotDagCidFinalized events raised by the ContractApi contract.
type ContractApiSnapshotDagCidFinalizedIterator struct {
	Event *ContractApiSnapshotDagCidFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiSnapshotDagCidFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiSnapshotDagCidFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiSnapshotDagCidFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiSnapshotDagCidFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiSnapshotDagCidFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiSnapshotDagCidFinalized represents a SnapshotDagCidFinalized event raised by the ContractApi contract.
type ContractApiSnapshotDagCidFinalized struct {
	ProjectId      string
	DAGBlockHeight *big.Int
	DagCid         string
	Timestamp      *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSnapshotDagCidFinalized is a free log retrieval operation binding the contract event 0x44b5b4670f388eef4ca390e24759bdfc52e752e486d069d0b109646ad44b7d2c.
//
// Solidity: event SnapshotDagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterSnapshotDagCidFinalized(opts *bind.FilterOpts, DAGBlockHeight []*big.Int, timestamp []*big.Int) (*ContractApiSnapshotDagCidFinalizedIterator, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotDagCidFinalized", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotDagCidFinalizedIterator{contract: _ContractApi.contract, event: "SnapshotDagCidFinalized", logs: logs, sub: sub}, nil
}

// WatchSnapshotDagCidFinalized is a free log subscription operation binding the contract event 0x44b5b4670f388eef4ca390e24759bdfc52e752e486d069d0b109646ad44b7d2c.
//
// Solidity: event SnapshotDagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchSnapshotDagCidFinalized(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotDagCidFinalized, DAGBlockHeight []*big.Int, timestamp []*big.Int) (event.Subscription, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotDagCidFinalized", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiSnapshotDagCidFinalized)
				if err := _ContractApi.contract.UnpackLog(event, "SnapshotDagCidFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSnapshotDagCidFinalized is a log parse operation binding the contract event 0x44b5b4670f388eef4ca390e24759bdfc52e752e486d069d0b109646ad44b7d2c.
//
// Solidity: event SnapshotDagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseSnapshotDagCidFinalized(log types.Log) (*ContractApiSnapshotDagCidFinalized, error) {
	event := new(ContractApiSnapshotDagCidFinalized)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshotDagCidFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiSnapshotDagCidSubmittedIterator is returned from FilterSnapshotDagCidSubmitted and is used to iterate over the raw logs and unpacked data for SnapshotDagCidSubmitted events raised by the ContractApi contract.
type ContractApiSnapshotDagCidSubmittedIterator struct {
	Event *ContractApiSnapshotDagCidSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiSnapshotDagCidSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiSnapshotDagCidSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiSnapshotDagCidSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiSnapshotDagCidSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiSnapshotDagCidSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiSnapshotDagCidSubmitted represents a SnapshotDagCidSubmitted event raised by the ContractApi contract.
type ContractApiSnapshotDagCidSubmitted struct {
	ProjectId      string
	DAGBlockHeight *big.Int
	DagCid         string
	Timestamp      *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSnapshotDagCidSubmitted is a free log retrieval operation binding the contract event 0xc4a7a58ee57b6e41ef76bd159f2358f00c321f9eb6e42b7bc1d4d6556ba67ea4.
//
// Solidity: event SnapshotDagCidSubmitted(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterSnapshotDagCidSubmitted(opts *bind.FilterOpts, DAGBlockHeight []*big.Int, timestamp []*big.Int) (*ContractApiSnapshotDagCidSubmittedIterator, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotDagCidSubmitted", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotDagCidSubmittedIterator{contract: _ContractApi.contract, event: "SnapshotDagCidSubmitted", logs: logs, sub: sub}, nil
}

// WatchSnapshotDagCidSubmitted is a free log subscription operation binding the contract event 0xc4a7a58ee57b6e41ef76bd159f2358f00c321f9eb6e42b7bc1d4d6556ba67ea4.
//
// Solidity: event SnapshotDagCidSubmitted(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchSnapshotDagCidSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotDagCidSubmitted, DAGBlockHeight []*big.Int, timestamp []*big.Int) (event.Subscription, error) {

	var DAGBlockHeightRule []interface{}
	for _, DAGBlockHeightItem := range DAGBlockHeight {
		DAGBlockHeightRule = append(DAGBlockHeightRule, DAGBlockHeightItem)
	}

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotDagCidSubmitted", DAGBlockHeightRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiSnapshotDagCidSubmitted)
				if err := _ContractApi.contract.UnpackLog(event, "SnapshotDagCidSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSnapshotDagCidSubmitted is a log parse operation binding the contract event 0xc4a7a58ee57b6e41ef76bd159f2358f00c321f9eb6e42b7bc1d4d6556ba67ea4.
//
// Solidity: event SnapshotDagCidSubmitted(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseSnapshotDagCidSubmitted(log types.Log) (*ContractApiSnapshotDagCidSubmitted, error) {
	event := new(ContractApiSnapshotDagCidSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshotDagCidSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiSnapshotFinalizedIterator is returned from FilterSnapshotFinalized and is used to iterate over the raw logs and unpacked data for SnapshotFinalized events raised by the ContractApi contract.
type ContractApiSnapshotFinalizedIterator struct {
	Event *ContractApiSnapshotFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiSnapshotFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiSnapshotFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiSnapshotFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiSnapshotFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiSnapshotFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiSnapshotFinalized represents a SnapshotFinalized event raised by the ContractApi contract.
type ContractApiSnapshotFinalized struct {
	EpochEnd    *big.Int
	ProjectId   string
	SnapshotCid string
	Timestamp   *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSnapshotFinalized is a free log retrieval operation binding the contract event 0xc4c418f2624540bbfeee5b4a52bc05e3d0372ce4d0baeb6103c80fedadc21ffb.
//
// Solidity: event SnapshotFinalized(uint256 epochEnd, string projectId, string snapshotCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterSnapshotFinalized(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiSnapshotFinalizedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotFinalized", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotFinalizedIterator{contract: _ContractApi.contract, event: "SnapshotFinalized", logs: logs, sub: sub}, nil
}

// WatchSnapshotFinalized is a free log subscription operation binding the contract event 0xc4c418f2624540bbfeee5b4a52bc05e3d0372ce4d0baeb6103c80fedadc21ffb.
//
// Solidity: event SnapshotFinalized(uint256 epochEnd, string projectId, string snapshotCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchSnapshotFinalized(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotFinalized, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotFinalized", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiSnapshotFinalized)
				if err := _ContractApi.contract.UnpackLog(event, "SnapshotFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSnapshotFinalized is a log parse operation binding the contract event 0xc4c418f2624540bbfeee5b4a52bc05e3d0372ce4d0baeb6103c80fedadc21ffb.
//
// Solidity: event SnapshotFinalized(uint256 epochEnd, string projectId, string snapshotCid, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseSnapshotFinalized(log types.Log) (*ContractApiSnapshotFinalized, error) {
	event := new(ContractApiSnapshotFinalized)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshotFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiSnapshotSubmittedIterator is returned from FilterSnapshotSubmitted and is used to iterate over the raw logs and unpacked data for SnapshotSubmitted events raised by the ContractApi contract.
type ContractApiSnapshotSubmittedIterator struct {
	Event *ContractApiSnapshotSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiSnapshotSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiSnapshotSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiSnapshotSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiSnapshotSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiSnapshotSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiSnapshotSubmitted represents a SnapshotSubmitted event raised by the ContractApi contract.
type ContractApiSnapshotSubmitted struct {
	SnapshotterAddr common.Address
	SnapshotCid     string
	EpochEnd        *big.Int
	ProjectId       string
	Timestamp       *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSnapshotSubmitted is a free log retrieval operation binding the contract event 0x2ad090680d8ecd2d9f1837e3a1df9c5ba943a1c047fe68cf5eaf69f88b7331e3.
//
// Solidity: event SnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) FilterSnapshotSubmitted(opts *bind.FilterOpts, timestamp []*big.Int) (*ContractApiSnapshotSubmittedIterator, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotSubmittedIterator{contract: _ContractApi.contract, event: "SnapshotSubmitted", logs: logs, sub: sub}, nil
}

// WatchSnapshotSubmitted is a free log subscription operation binding the contract event 0x2ad090680d8ecd2d9f1837e3a1df9c5ba943a1c047fe68cf5eaf69f88b7331e3.
//
// Solidity: event SnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) WatchSnapshotSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotSubmitted, timestamp []*big.Int) (event.Subscription, error) {

	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotSubmitted", timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiSnapshotSubmitted)
				if err := _ContractApi.contract.UnpackLog(event, "SnapshotSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSnapshotSubmitted is a log parse operation binding the contract event 0x2ad090680d8ecd2d9f1837e3a1df9c5ba943a1c047fe68cf5eaf69f88b7331e3.
//
// Solidity: event SnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 epochEnd, string projectId, uint256 indexed timestamp)
func (_ContractApi *ContractApiFilterer) ParseSnapshotSubmitted(log types.Log) (*ContractApiSnapshotSubmitted, error) {
	event := new(ContractApiSnapshotSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshotSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiSnapshotterAllowedIterator is returned from FilterSnapshotterAllowed and is used to iterate over the raw logs and unpacked data for SnapshotterAllowed events raised by the ContractApi contract.
type ContractApiSnapshotterAllowedIterator struct {
	Event *ContractApiSnapshotterAllowed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiSnapshotterAllowedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiSnapshotterAllowed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiSnapshotterAllowed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiSnapshotterAllowedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiSnapshotterAllowedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiSnapshotterAllowed represents a SnapshotterAllowed event raised by the ContractApi contract.
type ContractApiSnapshotterAllowed struct {
	SnapshotterAddress common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterSnapshotterAllowed is a free log retrieval operation binding the contract event 0xfcb9509c76f3ebc07afe1d348f7280fb6a952d80f78357319de7ce75a1350410.
//
// Solidity: event SnapshotterAllowed(address snapshotterAddress)
func (_ContractApi *ContractApiFilterer) FilterSnapshotterAllowed(opts *bind.FilterOpts) (*ContractApiSnapshotterAllowedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotterAllowed")
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotterAllowedIterator{contract: _ContractApi.contract, event: "SnapshotterAllowed", logs: logs, sub: sub}, nil
}

// WatchSnapshotterAllowed is a free log subscription operation binding the contract event 0xfcb9509c76f3ebc07afe1d348f7280fb6a952d80f78357319de7ce75a1350410.
//
// Solidity: event SnapshotterAllowed(address snapshotterAddress)
func (_ContractApi *ContractApiFilterer) WatchSnapshotterAllowed(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotterAllowed) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotterAllowed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiSnapshotterAllowed)
				if err := _ContractApi.contract.UnpackLog(event, "SnapshotterAllowed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSnapshotterAllowed is a log parse operation binding the contract event 0xfcb9509c76f3ebc07afe1d348f7280fb6a952d80f78357319de7ce75a1350410.
//
// Solidity: event SnapshotterAllowed(address snapshotterAddress)
func (_ContractApi *ContractApiFilterer) ParseSnapshotterAllowed(log types.Log) (*ContractApiSnapshotterAllowed, error) {
	event := new(ContractApiSnapshotterAllowed)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshotterAllowed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiSnapshotterRegisteredIterator is returned from FilterSnapshotterRegistered and is used to iterate over the raw logs and unpacked data for SnapshotterRegistered events raised by the ContractApi contract.
type ContractApiSnapshotterRegisteredIterator struct {
	Event *ContractApiSnapshotterRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiSnapshotterRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiSnapshotterRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiSnapshotterRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiSnapshotterRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiSnapshotterRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiSnapshotterRegistered represents a SnapshotterRegistered event raised by the ContractApi contract.
type ContractApiSnapshotterRegistered struct {
	SnapshotterAddr common.Address
	ProjectId       string
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSnapshotterRegistered is a free log retrieval operation binding the contract event 0x10af12c3369726967afdb89619a5c2b6ac5a43b59bf910796112bf60ddf88e3a.
//
// Solidity: event SnapshotterRegistered(address snapshotterAddr, string projectId)
func (_ContractApi *ContractApiFilterer) FilterSnapshotterRegistered(opts *bind.FilterOpts) (*ContractApiSnapshotterRegisteredIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotterRegistered")
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotterRegisteredIterator{contract: _ContractApi.contract, event: "SnapshotterRegistered", logs: logs, sub: sub}, nil
}

// WatchSnapshotterRegistered is a free log subscription operation binding the contract event 0x10af12c3369726967afdb89619a5c2b6ac5a43b59bf910796112bf60ddf88e3a.
//
// Solidity: event SnapshotterRegistered(address snapshotterAddr, string projectId)
func (_ContractApi *ContractApiFilterer) WatchSnapshotterRegistered(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotterRegistered) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotterRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiSnapshotterRegistered)
				if err := _ContractApi.contract.UnpackLog(event, "SnapshotterRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSnapshotterRegistered is a log parse operation binding the contract event 0x10af12c3369726967afdb89619a5c2b6ac5a43b59bf910796112bf60ddf88e3a.
//
// Solidity: event SnapshotterRegistered(address snapshotterAddr, string projectId)
func (_ContractApi *ContractApiFilterer) ParseSnapshotterRegistered(log types.Log) (*ContractApiSnapshotterRegistered, error) {
	event := new(ContractApiSnapshotterRegistered)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshotterRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiStateBuilderAllowedIterator is returned from FilterStateBuilderAllowed and is used to iterate over the raw logs and unpacked data for StateBuilderAllowed events raised by the ContractApi contract.
type ContractApiStateBuilderAllowedIterator struct {
	Event *ContractApiStateBuilderAllowed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractApiStateBuilderAllowedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiStateBuilderAllowed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractApiStateBuilderAllowed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractApiStateBuilderAllowedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiStateBuilderAllowedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiStateBuilderAllowed represents a StateBuilderAllowed event raised by the ContractApi contract.
type ContractApiStateBuilderAllowed struct {
	StateBuilderAddress common.Address
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterStateBuilderAllowed is a free log retrieval operation binding the contract event 0x24039cc8f0514f4700504c3e86aee62f4df8cffa388c2373d048ba2dec79af0e.
//
// Solidity: event StateBuilderAllowed(address stateBuilderAddress)
func (_ContractApi *ContractApiFilterer) FilterStateBuilderAllowed(opts *bind.FilterOpts) (*ContractApiStateBuilderAllowedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "StateBuilderAllowed")
	if err != nil {
		return nil, err
	}
	return &ContractApiStateBuilderAllowedIterator{contract: _ContractApi.contract, event: "StateBuilderAllowed", logs: logs, sub: sub}, nil
}

// WatchStateBuilderAllowed is a free log subscription operation binding the contract event 0x24039cc8f0514f4700504c3e86aee62f4df8cffa388c2373d048ba2dec79af0e.
//
// Solidity: event StateBuilderAllowed(address stateBuilderAddress)
func (_ContractApi *ContractApiFilterer) WatchStateBuilderAllowed(opts *bind.WatchOpts, sink chan<- *ContractApiStateBuilderAllowed) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "StateBuilderAllowed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiStateBuilderAllowed)
				if err := _ContractApi.contract.UnpackLog(event, "StateBuilderAllowed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStateBuilderAllowed is a log parse operation binding the contract event 0x24039cc8f0514f4700504c3e86aee62f4df8cffa388c2373d048ba2dec79af0e.
//
// Solidity: event StateBuilderAllowed(address stateBuilderAddress)
func (_ContractApi *ContractApiFilterer) ParseStateBuilderAllowed(log types.Log) (*ContractApiStateBuilderAllowed, error) {
	event := new(ContractApiStateBuilderAllowed)
	if err := _ContractApi.contract.UnpackLog(event, "StateBuilderAllowed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

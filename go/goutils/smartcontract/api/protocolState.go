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

// PowerloomProtocolStateRequest is an auto generated low-level Go binding around an user-defined struct.
type PowerloomProtocolStateRequest struct {
	Deadline    *big.Int
	SnapshotCid string
	EpochId     *big.Int
	ProjectId   string
}

// ContractApiMetaData contains all meta data concerning the ContractApi contract.
var ContractApiMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DelayedSnapshotSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"EpochReleased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"SnapshottersUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"validatorAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"ValidatorsUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"EPOCH_SIZE\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SOURCE_CHAIN_BLOCK_TIME\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SOURCE_CHAIN_ID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"name\":\"checkDynamicConsensusSnapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"epochInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blocknumber\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"name\":\"forceCompleteConsensusSnapshot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllSnapshotters\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProjects\",\"outputs\":[{\"internalType\":\"string[]\",\"name\":\"\",\"type\":\"string[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalSnapshotterCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getValidators\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxSnapshotsCid\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxSnapshotsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minSubmissionsForConsensus\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"projectFirstEpochId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"messageHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"recoverAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"releaseEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"snapshotStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"finalized\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"snapshotSubmissionWindow\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"snapshotsReceived\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"snapshotsReceivedCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"snapshotters\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"}],\"internalType\":\"structPowerloomProtocolState.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"submitSnapshot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minSubmissionsForConsensus\",\"type\":\"uint256\"}],\"name\":\"updateMinSnapshottersForConsensus\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newsnapshotSubmissionWindow\",\"type\":\"uint256\"}],\"name\":\"updateSnapshotSubmissionWindow\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_snapshotters\",\"type\":\"address[]\"},{\"internalType\":\"bool[]\",\"name\":\"_status\",\"type\":\"bool[]\"}],\"name\":\"updateSnapshotters\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_validators\",\"type\":\"address[]\"},{\"internalType\":\"bool[]\",\"name\":\"_status\",\"type\":\"bool[]\"}],\"name\":\"updateValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"}],\"internalType\":\"structPowerloomProtocolState.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"verify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ContractApiABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractApiMetaData.ABI instead.
var ContractApiABI = ContractApiMetaData.ABI

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

// EPOCHSIZE is a free data retrieval call binding the contract method 0x62656003.
//
// Solidity: function EPOCH_SIZE() view returns(uint8)
func (_ContractApi *ContractApiCaller) EPOCHSIZE(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "EPOCH_SIZE")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// EPOCHSIZE is a free data retrieval call binding the contract method 0x62656003.
//
// Solidity: function EPOCH_SIZE() view returns(uint8)
func (_ContractApi *ContractApiSession) EPOCHSIZE() (uint8, error) {
	return _ContractApi.Contract.EPOCHSIZE(&_ContractApi.CallOpts)
}

// EPOCHSIZE is a free data retrieval call binding the contract method 0x62656003.
//
// Solidity: function EPOCH_SIZE() view returns(uint8)
func (_ContractApi *ContractApiCallerSession) EPOCHSIZE() (uint8, error) {
	return _ContractApi.Contract.EPOCHSIZE(&_ContractApi.CallOpts)
}

// SOURCECHAINBLOCKTIME is a free data retrieval call binding the contract method 0x351b6155.
//
// Solidity: function SOURCE_CHAIN_BLOCK_TIME() view returns(uint256)
func (_ContractApi *ContractApiCaller) SOURCECHAINBLOCKTIME(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "SOURCE_CHAIN_BLOCK_TIME")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SOURCECHAINBLOCKTIME is a free data retrieval call binding the contract method 0x351b6155.
//
// Solidity: function SOURCE_CHAIN_BLOCK_TIME() view returns(uint256)
func (_ContractApi *ContractApiSession) SOURCECHAINBLOCKTIME() (*big.Int, error) {
	return _ContractApi.Contract.SOURCECHAINBLOCKTIME(&_ContractApi.CallOpts)
}

// SOURCECHAINBLOCKTIME is a free data retrieval call binding the contract method 0x351b6155.
//
// Solidity: function SOURCE_CHAIN_BLOCK_TIME() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SOURCECHAINBLOCKTIME() (*big.Int, error) {
	return _ContractApi.Contract.SOURCECHAINBLOCKTIME(&_ContractApi.CallOpts)
}

// SOURCECHAINID is a free data retrieval call binding the contract method 0x74be2150.
//
// Solidity: function SOURCE_CHAIN_ID() view returns(uint256)
func (_ContractApi *ContractApiCaller) SOURCECHAINID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "SOURCE_CHAIN_ID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SOURCECHAINID is a free data retrieval call binding the contract method 0x74be2150.
//
// Solidity: function SOURCE_CHAIN_ID() view returns(uint256)
func (_ContractApi *ContractApiSession) SOURCECHAINID() (*big.Int, error) {
	return _ContractApi.Contract.SOURCECHAINID(&_ContractApi.CallOpts)
}

// SOURCECHAINID is a free data retrieval call binding the contract method 0x74be2150.
//
// Solidity: function SOURCE_CHAIN_ID() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SOURCECHAINID() (*big.Int, error) {
	return _ContractApi.Contract.SOURCECHAINID(&_ContractApi.CallOpts)
}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochId) view returns(bool)
func (_ContractApi *ContractApiCaller) CheckDynamicConsensusSnapshot(opts *bind.CallOpts, projectId string, epochId *big.Int) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "checkDynamicConsensusSnapshot", projectId, epochId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochId) view returns(bool)
func (_ContractApi *ContractApiSession) CheckDynamicConsensusSnapshot(projectId string, epochId *big.Int) (bool, error) {
	return _ContractApi.Contract.CheckDynamicConsensusSnapshot(&_ContractApi.CallOpts, projectId, epochId)
}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochId) view returns(bool)
func (_ContractApi *ContractApiCallerSession) CheckDynamicConsensusSnapshot(projectId string, epochId *big.Int) (bool, error) {
	return _ContractApi.Contract.CheckDynamicConsensusSnapshot(&_ContractApi.CallOpts, projectId, epochId)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256 begin, uint256 end, uint256 epochId)
func (_ContractApi *ContractApiCaller) CurrentEpoch(opts *bind.CallOpts) (struct {
	Begin   *big.Int
	End     *big.Int
	EpochId *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "currentEpoch")

	outstruct := new(struct {
		Begin   *big.Int
		End     *big.Int
		EpochId *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Begin = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.End = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.EpochId = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256 begin, uint256 end, uint256 epochId)
func (_ContractApi *ContractApiSession) CurrentEpoch() (struct {
	Begin   *big.Int
	End     *big.Int
	EpochId *big.Int
}, error) {
	return _ContractApi.Contract.CurrentEpoch(&_ContractApi.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256 begin, uint256 end, uint256 epochId)
func (_ContractApi *ContractApiCallerSession) CurrentEpoch() (struct {
	Begin   *big.Int
	End     *big.Int
	EpochId *big.Int
}, error) {
	return _ContractApi.Contract.CurrentEpoch(&_ContractApi.CallOpts)
}

// EpochInfo is a free data retrieval call binding the contract method 0x3894228e.
//
// Solidity: function epochInfo(uint256 ) view returns(uint256 timestamp, uint256 blocknumber, uint256 epochEnd)
func (_ContractApi *ContractApiCaller) EpochInfo(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
	EpochEnd    *big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "epochInfo", arg0)

	outstruct := new(struct {
		Timestamp   *big.Int
		Blocknumber *big.Int
		EpochEnd    *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Timestamp = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Blocknumber = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.EpochEnd = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// EpochInfo is a free data retrieval call binding the contract method 0x3894228e.
//
// Solidity: function epochInfo(uint256 ) view returns(uint256 timestamp, uint256 blocknumber, uint256 epochEnd)
func (_ContractApi *ContractApiSession) EpochInfo(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
	EpochEnd    *big.Int
}, error) {
	return _ContractApi.Contract.EpochInfo(&_ContractApi.CallOpts, arg0)
}

// EpochInfo is a free data retrieval call binding the contract method 0x3894228e.
//
// Solidity: function epochInfo(uint256 ) view returns(uint256 timestamp, uint256 blocknumber, uint256 epochEnd)
func (_ContractApi *ContractApiCallerSession) EpochInfo(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
	EpochEnd    *big.Int
}, error) {
	return _ContractApi.Contract.EpochInfo(&_ContractApi.CallOpts, arg0)
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

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_ContractApi *ContractApiCaller) GetValidators(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getValidators")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_ContractApi *ContractApiSession) GetValidators() ([]common.Address, error) {
	return _ContractApi.Contract.GetValidators(&_ContractApi.CallOpts)
}

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_ContractApi *ContractApiCallerSession) GetValidators() ([]common.Address, error) {
	return _ContractApi.Contract.GetValidators(&_ContractApi.CallOpts)
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

// ProjectFirstEpochId is a free data retrieval call binding the contract method 0xfa30dbe0.
//
// Solidity: function projectFirstEpochId(string ) view returns(uint256)
func (_ContractApi *ContractApiCaller) ProjectFirstEpochId(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "projectFirstEpochId", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProjectFirstEpochId is a free data retrieval call binding the contract method 0xfa30dbe0.
//
// Solidity: function projectFirstEpochId(string ) view returns(uint256)
func (_ContractApi *ContractApiSession) ProjectFirstEpochId(arg0 string) (*big.Int, error) {
	return _ContractApi.Contract.ProjectFirstEpochId(&_ContractApi.CallOpts, arg0)
}

// ProjectFirstEpochId is a free data retrieval call binding the contract method 0xfa30dbe0.
//
// Solidity: function projectFirstEpochId(string ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) ProjectFirstEpochId(arg0 string) (*big.Int, error) {
	return _ContractApi.Contract.ProjectFirstEpochId(&_ContractApi.CallOpts, arg0)
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

// Verify is a free data retrieval call binding the contract method 0xc22e1a53.
//
// Solidity: function verify(string snapshotCid, uint256 epochId, string projectId, (uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiCaller) Verify(opts *bind.CallOpts, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "verify", snapshotCid, epochId, projectId, request, signature, signer)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Verify is a free data retrieval call binding the contract method 0xc22e1a53.
//
// Solidity: function verify(string snapshotCid, uint256 epochId, string projectId, (uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiSession) Verify(snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	return _ContractApi.Contract.Verify(&_ContractApi.CallOpts, snapshotCid, epochId, projectId, request, signature, signer)
}

// Verify is a free data retrieval call binding the contract method 0xc22e1a53.
//
// Solidity: function verify(string snapshotCid, uint256 epochId, string projectId, (uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiCallerSession) Verify(snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	return _ContractApi.Contract.Verify(&_ContractApi.CallOpts, snapshotCid, epochId, projectId, request, signature, signer)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochId) returns()
func (_ContractApi *ContractApiTransactor) ForceCompleteConsensusSnapshot(opts *bind.TransactOpts, projectId string, epochId *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "forceCompleteConsensusSnapshot", projectId, epochId)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochId) returns()
func (_ContractApi *ContractApiSession) ForceCompleteConsensusSnapshot(projectId string, epochId *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceCompleteConsensusSnapshot(&_ContractApi.TransactOpts, projectId, epochId)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochId) returns()
func (_ContractApi *ContractApiTransactorSession) ForceCompleteConsensusSnapshot(projectId string, epochId *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceCompleteConsensusSnapshot(&_ContractApi.TransactOpts, projectId, epochId)
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

// SubmitSnapshot is a paid mutator transaction binding the contract method 0xae95eb50.
//
// Solidity: function submitSnapshot(string snapshotCid, uint256 epochId, string projectId, (uint256,string,uint256,string) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactor) SubmitSnapshot(opts *bind.TransactOpts, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "submitSnapshot", snapshotCid, epochId, projectId, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0xae95eb50.
//
// Solidity: function submitSnapshot(string snapshotCid, uint256 epochId, string projectId, (uint256,string,uint256,string) request, bytes signature) returns()
func (_ContractApi *ContractApiSession) SubmitSnapshot(snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitSnapshot(&_ContractApi.TransactOpts, snapshotCid, epochId, projectId, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0xae95eb50.
//
// Solidity: function submitSnapshot(string snapshotCid, uint256 epochId, string projectId, (uint256,string,uint256,string) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactorSession) SubmitSnapshot(snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitSnapshot(&_ContractApi.TransactOpts, snapshotCid, epochId, projectId, request, signature)
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

// UpdateSnapshotters is a paid mutator transaction binding the contract method 0xb8cf6fee.
//
// Solidity: function updateSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ContractApi *ContractApiTransactor) UpdateSnapshotters(opts *bind.TransactOpts, _snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateSnapshotters", _snapshotters, _status)
}

// UpdateSnapshotters is a paid mutator transaction binding the contract method 0xb8cf6fee.
//
// Solidity: function updateSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ContractApi *ContractApiSession) UpdateSnapshotters(_snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateSnapshotters(&_ContractApi.TransactOpts, _snapshotters, _status)
}

// UpdateSnapshotters is a paid mutator transaction binding the contract method 0xb8cf6fee.
//
// Solidity: function updateSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateSnapshotters(_snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateSnapshotters(&_ContractApi.TransactOpts, _snapshotters, _status)
}

// UpdateValidators is a paid mutator transaction binding the contract method 0x1b0b3ae3.
//
// Solidity: function updateValidators(address[] _validators, bool[] _status) returns()
func (_ContractApi *ContractApiTransactor) UpdateValidators(opts *bind.TransactOpts, _validators []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateValidators", _validators, _status)
}

// UpdateValidators is a paid mutator transaction binding the contract method 0x1b0b3ae3.
//
// Solidity: function updateValidators(address[] _validators, bool[] _status) returns()
func (_ContractApi *ContractApiSession) UpdateValidators(_validators []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateValidators(&_ContractApi.TransactOpts, _validators, _status)
}

// UpdateValidators is a paid mutator transaction binding the contract method 0x1b0b3ae3.
//
// Solidity: function updateValidators(address[] _validators, bool[] _status) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateValidators(_validators []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateValidators(&_ContractApi.TransactOpts, _validators, _status)
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
	EpochId         *big.Int
	ProjectId       string
	Timestamp       *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterDelayedSnapshotSubmitted is a free log retrieval operation binding the contract event 0xa4b1762053ee970f50692b6936d4e58a9a01291449e4da16bdf758891c8de752.
//
// Solidity: event DelayedSnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) FilterDelayedSnapshotSubmitted(opts *bind.FilterOpts, epochId []*big.Int) (*ContractApiDelayedSnapshotSubmittedIterator, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "DelayedSnapshotSubmitted", epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiDelayedSnapshotSubmittedIterator{contract: _ContractApi.contract, event: "DelayedSnapshotSubmitted", logs: logs, sub: sub}, nil
}

// WatchDelayedSnapshotSubmitted is a free log subscription operation binding the contract event 0xa4b1762053ee970f50692b6936d4e58a9a01291449e4da16bdf758891c8de752.
//
// Solidity: event DelayedSnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) WatchDelayedSnapshotSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiDelayedSnapshotSubmitted, epochId []*big.Int) (event.Subscription, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "DelayedSnapshotSubmitted", epochIdRule)
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
// Solidity: event DelayedSnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
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
	EpochId   *big.Int
	Begin     *big.Int
	End       *big.Int
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEpochReleased is a free log retrieval operation binding the contract event 0x108f87075a74f81fa2271fdf9fc0883a1811431182601fc65d24513970336640.
//
// Solidity: event EpochReleased(uint256 indexed epochId, uint256 begin, uint256 end, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) FilterEpochReleased(opts *bind.FilterOpts, epochId []*big.Int) (*ContractApiEpochReleasedIterator, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "EpochReleased", epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiEpochReleasedIterator{contract: _ContractApi.contract, event: "EpochReleased", logs: logs, sub: sub}, nil
}

// WatchEpochReleased is a free log subscription operation binding the contract event 0x108f87075a74f81fa2271fdf9fc0883a1811431182601fc65d24513970336640.
//
// Solidity: event EpochReleased(uint256 indexed epochId, uint256 begin, uint256 end, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) WatchEpochReleased(opts *bind.WatchOpts, sink chan<- *ContractApiEpochReleased, epochId []*big.Int) (event.Subscription, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "EpochReleased", epochIdRule)
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

// ParseEpochReleased is a log parse operation binding the contract event 0x108f87075a74f81fa2271fdf9fc0883a1811431182601fc65d24513970336640.
//
// Solidity: event EpochReleased(uint256 indexed epochId, uint256 begin, uint256 end, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) ParseEpochReleased(log types.Log) (*ContractApiEpochReleased, error) {
	event := new(ContractApiEpochReleased)
	if err := _ContractApi.contract.UnpackLog(event, "EpochReleased", log); err != nil {
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
	EpochId     *big.Int
	EpochEnd    *big.Int
	ProjectId   string
	SnapshotCid string
	Timestamp   *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSnapshotFinalized is a free log retrieval operation binding the contract event 0xe5231a68c59ef23c90b7da4209eae4c795477f0d5dcfa14a612ea96f69a18e15.
//
// Solidity: event SnapshotFinalized(uint256 indexed epochId, uint256 epochEnd, string projectId, string snapshotCid, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) FilterSnapshotFinalized(opts *bind.FilterOpts, epochId []*big.Int) (*ContractApiSnapshotFinalizedIterator, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotFinalized", epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotFinalizedIterator{contract: _ContractApi.contract, event: "SnapshotFinalized", logs: logs, sub: sub}, nil
}

// WatchSnapshotFinalized is a free log subscription operation binding the contract event 0xe5231a68c59ef23c90b7da4209eae4c795477f0d5dcfa14a612ea96f69a18e15.
//
// Solidity: event SnapshotFinalized(uint256 indexed epochId, uint256 epochEnd, string projectId, string snapshotCid, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) WatchSnapshotFinalized(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotFinalized, epochId []*big.Int) (event.Subscription, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotFinalized", epochIdRule)
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

// ParseSnapshotFinalized is a log parse operation binding the contract event 0xe5231a68c59ef23c90b7da4209eae4c795477f0d5dcfa14a612ea96f69a18e15.
//
// Solidity: event SnapshotFinalized(uint256 indexed epochId, uint256 epochEnd, string projectId, string snapshotCid, uint256 timestamp)
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
	EpochId         *big.Int
	ProjectId       string
	Timestamp       *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSnapshotSubmitted is a free log retrieval operation binding the contract event 0x2ad090680d8ecd2d9f1837e3a1df9c5ba943a1c047fe68cf5eaf69f88b7331e3.
//
// Solidity: event SnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) FilterSnapshotSubmitted(opts *bind.FilterOpts, epochId []*big.Int) (*ContractApiSnapshotSubmittedIterator, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotSubmitted", epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotSubmittedIterator{contract: _ContractApi.contract, event: "SnapshotSubmitted", logs: logs, sub: sub}, nil
}

// WatchSnapshotSubmitted is a free log subscription operation binding the contract event 0x2ad090680d8ecd2d9f1837e3a1df9c5ba943a1c047fe68cf5eaf69f88b7331e3.
//
// Solidity: event SnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) WatchSnapshotSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotSubmitted, epochId []*big.Int) (event.Subscription, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotSubmitted", epochIdRule)
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
// Solidity: event SnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) ParseSnapshotSubmitted(log types.Log) (*ContractApiSnapshotSubmitted, error) {
	event := new(ContractApiSnapshotSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshotSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiSnapshottersUpdatedIterator is returned from FilterSnapshottersUpdated and is used to iterate over the raw logs and unpacked data for SnapshottersUpdated events raised by the ContractApi contract.
type ContractApiSnapshottersUpdatedIterator struct {
	Event *ContractApiSnapshottersUpdated // Event containing the contract specifics and raw log

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
func (it *ContractApiSnapshottersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiSnapshottersUpdated)
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
		it.Event = new(ContractApiSnapshottersUpdated)
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
func (it *ContractApiSnapshottersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiSnapshottersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiSnapshottersUpdated represents a SnapshottersUpdated event raised by the ContractApi contract.
type ContractApiSnapshottersUpdated struct {
	SnapshotterAddress common.Address
	Allowed            bool
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterSnapshottersUpdated is a free log retrieval operation binding the contract event 0x4c765465b6151ce64f545c3d8a4a500d6c4f5ca0a25d17f4db5e8d76f72c37da.
//
// Solidity: event SnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) FilterSnapshottersUpdated(opts *bind.FilterOpts) (*ContractApiSnapshottersUpdatedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshottersUpdatedIterator{contract: _ContractApi.contract, event: "SnapshottersUpdated", logs: logs, sub: sub}, nil
}

// WatchSnapshottersUpdated is a free log subscription operation binding the contract event 0x4c765465b6151ce64f545c3d8a4a500d6c4f5ca0a25d17f4db5e8d76f72c37da.
//
// Solidity: event SnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) WatchSnapshottersUpdated(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshottersUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiSnapshottersUpdated)
				if err := _ContractApi.contract.UnpackLog(event, "SnapshottersUpdated", log); err != nil {
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

// ParseSnapshottersUpdated is a log parse operation binding the contract event 0x4c765465b6151ce64f545c3d8a4a500d6c4f5ca0a25d17f4db5e8d76f72c37da.
//
// Solidity: event SnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) ParseSnapshottersUpdated(log types.Log) (*ContractApiSnapshottersUpdated, error) {
	event := new(ContractApiSnapshottersUpdated)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshottersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiValidatorsUpdatedIterator is returned from FilterValidatorsUpdated and is used to iterate over the raw logs and unpacked data for ValidatorsUpdated events raised by the ContractApi contract.
type ContractApiValidatorsUpdatedIterator struct {
	Event *ContractApiValidatorsUpdated // Event containing the contract specifics and raw log

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
func (it *ContractApiValidatorsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiValidatorsUpdated)
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
		it.Event = new(ContractApiValidatorsUpdated)
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
func (it *ContractApiValidatorsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiValidatorsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiValidatorsUpdated represents a ValidatorsUpdated event raised by the ContractApi contract.
type ContractApiValidatorsUpdated struct {
	ValidatorAddress common.Address
	Allowed          bool
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterValidatorsUpdated is a free log retrieval operation binding the contract event 0x7f3079c058f3e3dee87048158309898b46e9741ff53b6c7a3afac7c370649afc.
//
// Solidity: event ValidatorsUpdated(address validatorAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) FilterValidatorsUpdated(opts *bind.FilterOpts) (*ContractApiValidatorsUpdatedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "ValidatorsUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractApiValidatorsUpdatedIterator{contract: _ContractApi.contract, event: "ValidatorsUpdated", logs: logs, sub: sub}, nil
}

// WatchValidatorsUpdated is a free log subscription operation binding the contract event 0x7f3079c058f3e3dee87048158309898b46e9741ff53b6c7a3afac7c370649afc.
//
// Solidity: event ValidatorsUpdated(address validatorAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) WatchValidatorsUpdated(opts *bind.WatchOpts, sink chan<- *ContractApiValidatorsUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "ValidatorsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiValidatorsUpdated)
				if err := _ContractApi.contract.UnpackLog(event, "ValidatorsUpdated", log); err != nil {
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

// ParseValidatorsUpdated is a log parse operation binding the contract event 0x7f3079c058f3e3dee87048158309898b46e9741ff53b6c7a3afac7c370649afc.
//
// Solidity: event ValidatorsUpdated(address validatorAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) ParseValidatorsUpdated(log types.Log) (*ContractApiValidatorsUpdated, error) {
	event := new(ContractApiValidatorsUpdated)
	if err := _ContractApi.contract.UnpackLog(event, "ValidatorsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

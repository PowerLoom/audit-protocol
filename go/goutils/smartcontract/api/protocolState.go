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
	SlotId      *big.Int
	Deadline    *big.Int
	SnapshotCid string
	EpochId     *big.Int
	ProjectId   string
}

// PowerloomProtocolStateSlotInfo is an auto generated low-level Go binding around an user-defined struct.
type PowerloomProtocolStateSlotInfo struct {
	SlotId                  *big.Int
	SnapshotterAddress      common.Address
	TimeSlot                *big.Int
	RewardPoints            *big.Int
	CurrentStreak           *big.Int
	CurrentStreakBonus      *big.Int
	CurrentDaySnapshotCount *big.Int
}

// ContractApiMetaData contains all meta data concerning the ContractApi contract.
var ContractApiMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"slotIds\",\"type\":\"uint256[]\"},{\"internalType\":\"address[]\",\"name\":\"snapshotterAddresses\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"timeSlots\",\"type\":\"uint256[]\"}],\"name\":\"assignSnapshotterToSlotsWithTimeSlot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeSlot\",\"type\":\"uint256\"}],\"name\":\"assignSnapshotterToSlotWithTimeSlot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"epochSize\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"sourceChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"sourceChainBlockTime\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"useBlockNumberAsEpochId\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"slotsPerDay\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"InvalidShortString\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"str\",\"type\":\"string\"}],\"name\":\"StringTooLong\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"dayId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DailyTaskCompletedEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"dayId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DayStartedEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DelayedSnapshotSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EIP712DomainChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"EpochReleased\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"name\":\"forceCompleteConsensusSnapshot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"forceSkipEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_dayCounter\",\"type\":\"uint256\"}],\"name\":\"loadCurrentDay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeSlot\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardPoints\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentStreak\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentDaySnapshotCount\",\"type\":\"uint256\"}],\"name\":\"loadSlotInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"dayId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"snapshotCount\",\"type\":\"uint256\"}],\"name\":\"loadSlotSubmissions\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectType\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"enableEpochId\",\"type\":\"uint256\"}],\"name\":\"ProjectTypeUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"releaseEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"finalizedSnapshotCount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalReceivedCount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotSubmitted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"}],\"internalType\":\"structPowerloomProtocolState.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"submitSnapshot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"toggleRewards\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"toggleTimeSlotCheck\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_projectType\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"_status\",\"type\":\"bool\"}],\"name\":\"updateAllowedProjectType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dailySnapshotQuota\",\"type\":\"uint256\"}],\"name\":\"updateDailySnapshotQuota\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_snapshotters\",\"type\":\"address[]\"},{\"internalType\":\"bool[]\",\"name\":\"_status\",\"type\":\"bool[]\"}],\"name\":\"updateMasterSnapshotters\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minSubmissionsForConsensus\",\"type\":\"uint256\"}],\"name\":\"updateMinSnapshottersForConsensus\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newrewardBasePoints\",\"type\":\"uint256\"}],\"name\":\"updaterewardBasePoints\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newsnapshotSubmissionWindow\",\"type\":\"uint256\"}],\"name\":\"updateSnapshotSubmissionWindow\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newstreakBonusPoints\",\"type\":\"uint256\"}],\"name\":\"updatestreakBonusPoints\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_validators\",\"type\":\"address[]\"},{\"internalType\":\"bool[]\",\"name\":\"_status\",\"type\":\"bool[]\"}],\"name\":\"updateValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"validatorAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"ValidatorsUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"allSnapshottersUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"forceStartDay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"masterSnapshottersUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"allowedProjectTypes\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"allSnapshotters\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"name\":\"checkDynamicConsensusSnapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"day\",\"type\":\"uint256\"}],\"name\":\"checkSlotTaskStatusForDay\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DailySnapshotQuota\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dayCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DeploymentBlockNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"EPOCH_SIZE\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"epochInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blocknumber\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"epochsInADay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMasterSnapshotters\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"}],\"name\":\"getSlotInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeSlot\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardPoints\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentStreak\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentStreakBonus\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentDaySnapshotCount\",\"type\":\"uint256\"}],\"internalType\":\"structPowerloomProtocolState.SlotInfo\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"}],\"name\":\"getSlotStreak\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_slotId\",\"type\":\"uint256\"}],\"name\":\"getSnapshotterTimeSlot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalMasterSnapshotterCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalSnapshotterCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getValidators\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"lastFinalizedSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"lastSnapshotterAddressUpdate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"masterSnapshotters\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxSnapshotsCid\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxSnapshotsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minSubmissionsForConsensus\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"projectFirstEpochId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"messageHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"recoverAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardBasePoints\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardsEnabled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"slotCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slotRewardPoints\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SLOTS_PER_DAY\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slotSnapshotterMapping\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slotStreakCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slotSubmissionCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"snapshotsReceived\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"snapshotsReceivedCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"snapshotsReceivedSlot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"snapshotStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"finalized\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"snapshotSubmissionWindow\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SOURCE_CHAIN_BLOCK_TIME\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SOURCE_CHAIN_ID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"streakBonusPoints\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeSlotCheck\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"timeSlotPreference\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"totalSnapshotsReceived\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"USE_BLOCK_NUMBER_AS_EPOCH_ID\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"}],\"internalType\":\"structPowerloomProtocolState.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"verify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
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

// DailySnapshotQuota is a free data retrieval call binding the contract method 0xc7c7077e.
//
// Solidity: function DailySnapshotQuota() view returns(uint256)
func (_ContractApi *ContractApiCaller) DailySnapshotQuota(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "DailySnapshotQuota")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DailySnapshotQuota is a free data retrieval call binding the contract method 0xc7c7077e.
//
// Solidity: function DailySnapshotQuota() view returns(uint256)
func (_ContractApi *ContractApiSession) DailySnapshotQuota() (*big.Int, error) {
	return _ContractApi.Contract.DailySnapshotQuota(&_ContractApi.CallOpts)
}

// DailySnapshotQuota is a free data retrieval call binding the contract method 0xc7c7077e.
//
// Solidity: function DailySnapshotQuota() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) DailySnapshotQuota() (*big.Int, error) {
	return _ContractApi.Contract.DailySnapshotQuota(&_ContractApi.CallOpts)
}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xb1288f71.
//
// Solidity: function DeploymentBlockNumber() view returns(uint256)
func (_ContractApi *ContractApiCaller) DeploymentBlockNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "DeploymentBlockNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xb1288f71.
//
// Solidity: function DeploymentBlockNumber() view returns(uint256)
func (_ContractApi *ContractApiSession) DeploymentBlockNumber() (*big.Int, error) {
	return _ContractApi.Contract.DeploymentBlockNumber(&_ContractApi.CallOpts)
}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xb1288f71.
//
// Solidity: function DeploymentBlockNumber() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) DeploymentBlockNumber() (*big.Int, error) {
	return _ContractApi.Contract.DeploymentBlockNumber(&_ContractApi.CallOpts)
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

// SLOTSPERDAY is a free data retrieval call binding the contract method 0xe479b88d.
//
// Solidity: function SLOTS_PER_DAY() view returns(uint256)
func (_ContractApi *ContractApiCaller) SLOTSPERDAY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "SLOTS_PER_DAY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SLOTSPERDAY is a free data retrieval call binding the contract method 0xe479b88d.
//
// Solidity: function SLOTS_PER_DAY() view returns(uint256)
func (_ContractApi *ContractApiSession) SLOTSPERDAY() (*big.Int, error) {
	return _ContractApi.Contract.SLOTSPERDAY(&_ContractApi.CallOpts)
}

// SLOTSPERDAY is a free data retrieval call binding the contract method 0xe479b88d.
//
// Solidity: function SLOTS_PER_DAY() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SLOTSPERDAY() (*big.Int, error) {
	return _ContractApi.Contract.SLOTSPERDAY(&_ContractApi.CallOpts)
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

// USEBLOCKNUMBERASEPOCHID is a free data retrieval call binding the contract method 0x2d46247b.
//
// Solidity: function USE_BLOCK_NUMBER_AS_EPOCH_ID() view returns(bool)
func (_ContractApi *ContractApiCaller) USEBLOCKNUMBERASEPOCHID(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "USE_BLOCK_NUMBER_AS_EPOCH_ID")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// USEBLOCKNUMBERASEPOCHID is a free data retrieval call binding the contract method 0x2d46247b.
//
// Solidity: function USE_BLOCK_NUMBER_AS_EPOCH_ID() view returns(bool)
func (_ContractApi *ContractApiSession) USEBLOCKNUMBERASEPOCHID() (bool, error) {
	return _ContractApi.Contract.USEBLOCKNUMBERASEPOCHID(&_ContractApi.CallOpts)
}

// USEBLOCKNUMBERASEPOCHID is a free data retrieval call binding the contract method 0x2d46247b.
//
// Solidity: function USE_BLOCK_NUMBER_AS_EPOCH_ID() view returns(bool)
func (_ContractApi *ContractApiCallerSession) USEBLOCKNUMBERASEPOCHID() (bool, error) {
	return _ContractApi.Contract.USEBLOCKNUMBERASEPOCHID(&_ContractApi.CallOpts)
}

// AllSnapshotters is a free data retrieval call binding the contract method 0x3d15d0f4.
//
// Solidity: function allSnapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiCaller) AllSnapshotters(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "allSnapshotters", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllSnapshotters is a free data retrieval call binding the contract method 0x3d15d0f4.
//
// Solidity: function allSnapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiSession) AllSnapshotters(arg0 common.Address) (bool, error) {
	return _ContractApi.Contract.AllSnapshotters(&_ContractApi.CallOpts, arg0)
}

// AllSnapshotters is a free data retrieval call binding the contract method 0x3d15d0f4.
//
// Solidity: function allSnapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiCallerSession) AllSnapshotters(arg0 common.Address) (bool, error) {
	return _ContractApi.Contract.AllSnapshotters(&_ContractApi.CallOpts, arg0)
}

// AllowedProjectTypes is a free data retrieval call binding the contract method 0x04d76d59.
//
// Solidity: function allowedProjectTypes(string ) view returns(bool)
func (_ContractApi *ContractApiCaller) AllowedProjectTypes(opts *bind.CallOpts, arg0 string) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "allowedProjectTypes", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllowedProjectTypes is a free data retrieval call binding the contract method 0x04d76d59.
//
// Solidity: function allowedProjectTypes(string ) view returns(bool)
func (_ContractApi *ContractApiSession) AllowedProjectTypes(arg0 string) (bool, error) {
	return _ContractApi.Contract.AllowedProjectTypes(&_ContractApi.CallOpts, arg0)
}

// AllowedProjectTypes is a free data retrieval call binding the contract method 0x04d76d59.
//
// Solidity: function allowedProjectTypes(string ) view returns(bool)
func (_ContractApi *ContractApiCallerSession) AllowedProjectTypes(arg0 string) (bool, error) {
	return _ContractApi.Contract.AllowedProjectTypes(&_ContractApi.CallOpts, arg0)
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

// CheckSlotTaskStatusForDay is a free data retrieval call binding the contract method 0xd1dd6ddd.
//
// Solidity: function checkSlotTaskStatusForDay(uint256 slotId, uint256 day) view returns(bool)
func (_ContractApi *ContractApiCaller) CheckSlotTaskStatusForDay(opts *bind.CallOpts, slotId *big.Int, day *big.Int) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "checkSlotTaskStatusForDay", slotId, day)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckSlotTaskStatusForDay is a free data retrieval call binding the contract method 0xd1dd6ddd.
//
// Solidity: function checkSlotTaskStatusForDay(uint256 slotId, uint256 day) view returns(bool)
func (_ContractApi *ContractApiSession) CheckSlotTaskStatusForDay(slotId *big.Int, day *big.Int) (bool, error) {
	return _ContractApi.Contract.CheckSlotTaskStatusForDay(&_ContractApi.CallOpts, slotId, day)
}

// CheckSlotTaskStatusForDay is a free data retrieval call binding the contract method 0xd1dd6ddd.
//
// Solidity: function checkSlotTaskStatusForDay(uint256 slotId, uint256 day) view returns(bool)
func (_ContractApi *ContractApiCallerSession) CheckSlotTaskStatusForDay(slotId *big.Int, day *big.Int) (bool, error) {
	return _ContractApi.Contract.CheckSlotTaskStatusForDay(&_ContractApi.CallOpts, slotId, day)
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

// DayCounter is a free data retrieval call binding the contract method 0x99332c5e.
//
// Solidity: function dayCounter() view returns(uint256)
func (_ContractApi *ContractApiCaller) DayCounter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "dayCounter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DayCounter is a free data retrieval call binding the contract method 0x99332c5e.
//
// Solidity: function dayCounter() view returns(uint256)
func (_ContractApi *ContractApiSession) DayCounter() (*big.Int, error) {
	return _ContractApi.Contract.DayCounter(&_ContractApi.CallOpts)
}

// DayCounter is a free data retrieval call binding the contract method 0x99332c5e.
//
// Solidity: function dayCounter() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) DayCounter() (*big.Int, error) {
	return _ContractApi.Contract.DayCounter(&_ContractApi.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_ContractApi *ContractApiCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_ContractApi *ContractApiSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _ContractApi.Contract.Eip712Domain(&_ContractApi.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_ContractApi *ContractApiCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _ContractApi.Contract.Eip712Domain(&_ContractApi.CallOpts)
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

// EpochsInADay is a free data retrieval call binding the contract method 0xe3042b17.
//
// Solidity: function epochsInADay() view returns(uint256)
func (_ContractApi *ContractApiCaller) EpochsInADay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "epochsInADay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EpochsInADay is a free data retrieval call binding the contract method 0xe3042b17.
//
// Solidity: function epochsInADay() view returns(uint256)
func (_ContractApi *ContractApiSession) EpochsInADay() (*big.Int, error) {
	return _ContractApi.Contract.EpochsInADay(&_ContractApi.CallOpts)
}

// EpochsInADay is a free data retrieval call binding the contract method 0xe3042b17.
//
// Solidity: function epochsInADay() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) EpochsInADay() (*big.Int, error) {
	return _ContractApi.Contract.EpochsInADay(&_ContractApi.CallOpts)
}

// GetMasterSnapshotters is a free data retrieval call binding the contract method 0x90110313.
//
// Solidity: function getMasterSnapshotters() view returns(address[])
func (_ContractApi *ContractApiCaller) GetMasterSnapshotters(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getMasterSnapshotters")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetMasterSnapshotters is a free data retrieval call binding the contract method 0x90110313.
//
// Solidity: function getMasterSnapshotters() view returns(address[])
func (_ContractApi *ContractApiSession) GetMasterSnapshotters() ([]common.Address, error) {
	return _ContractApi.Contract.GetMasterSnapshotters(&_ContractApi.CallOpts)
}

// GetMasterSnapshotters is a free data retrieval call binding the contract method 0x90110313.
//
// Solidity: function getMasterSnapshotters() view returns(address[])
func (_ContractApi *ContractApiCallerSession) GetMasterSnapshotters() ([]common.Address, error) {
	return _ContractApi.Contract.GetMasterSnapshotters(&_ContractApi.CallOpts)
}

// GetSlotInfo is a free data retrieval call binding the contract method 0xbe20f9ac.
//
// Solidity: function getSlotInfo(uint256 slotId) view returns((uint256,address,uint256,uint256,uint256,uint256,uint256))
func (_ContractApi *ContractApiCaller) GetSlotInfo(opts *bind.CallOpts, slotId *big.Int) (PowerloomProtocolStateSlotInfo, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getSlotInfo", slotId)

	if err != nil {
		return *new(PowerloomProtocolStateSlotInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(PowerloomProtocolStateSlotInfo)).(*PowerloomProtocolStateSlotInfo)

	return out0, err

}

// GetSlotInfo is a free data retrieval call binding the contract method 0xbe20f9ac.
//
// Solidity: function getSlotInfo(uint256 slotId) view returns((uint256,address,uint256,uint256,uint256,uint256,uint256))
func (_ContractApi *ContractApiSession) GetSlotInfo(slotId *big.Int) (PowerloomProtocolStateSlotInfo, error) {
	return _ContractApi.Contract.GetSlotInfo(&_ContractApi.CallOpts, slotId)
}

// GetSlotInfo is a free data retrieval call binding the contract method 0xbe20f9ac.
//
// Solidity: function getSlotInfo(uint256 slotId) view returns((uint256,address,uint256,uint256,uint256,uint256,uint256))
func (_ContractApi *ContractApiCallerSession) GetSlotInfo(slotId *big.Int) (PowerloomProtocolStateSlotInfo, error) {
	return _ContractApi.Contract.GetSlotInfo(&_ContractApi.CallOpts, slotId)
}

// GetSlotStreak is a free data retrieval call binding the contract method 0x2252c7b8.
//
// Solidity: function getSlotStreak(uint256 slotId) view returns(uint256)
func (_ContractApi *ContractApiCaller) GetSlotStreak(opts *bind.CallOpts, slotId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getSlotStreak", slotId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlotStreak is a free data retrieval call binding the contract method 0x2252c7b8.
//
// Solidity: function getSlotStreak(uint256 slotId) view returns(uint256)
func (_ContractApi *ContractApiSession) GetSlotStreak(slotId *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.GetSlotStreak(&_ContractApi.CallOpts, slotId)
}

// GetSlotStreak is a free data retrieval call binding the contract method 0x2252c7b8.
//
// Solidity: function getSlotStreak(uint256 slotId) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) GetSlotStreak(slotId *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.GetSlotStreak(&_ContractApi.CallOpts, slotId)
}

// GetSnapshotterTimeSlot is a free data retrieval call binding the contract method 0x2db9ace2.
//
// Solidity: function getSnapshotterTimeSlot(uint256 _slotId) view returns(uint256)
func (_ContractApi *ContractApiCaller) GetSnapshotterTimeSlot(opts *bind.CallOpts, _slotId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getSnapshotterTimeSlot", _slotId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSnapshotterTimeSlot is a free data retrieval call binding the contract method 0x2db9ace2.
//
// Solidity: function getSnapshotterTimeSlot(uint256 _slotId) view returns(uint256)
func (_ContractApi *ContractApiSession) GetSnapshotterTimeSlot(_slotId *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.GetSnapshotterTimeSlot(&_ContractApi.CallOpts, _slotId)
}

// GetSnapshotterTimeSlot is a free data retrieval call binding the contract method 0x2db9ace2.
//
// Solidity: function getSnapshotterTimeSlot(uint256 _slotId) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) GetSnapshotterTimeSlot(_slotId *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.GetSnapshotterTimeSlot(&_ContractApi.CallOpts, _slotId)
}

// GetTotalMasterSnapshotterCount is a free data retrieval call binding the contract method 0xd551672b.
//
// Solidity: function getTotalMasterSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiCaller) GetTotalMasterSnapshotterCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "getTotalMasterSnapshotterCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalMasterSnapshotterCount is a free data retrieval call binding the contract method 0xd551672b.
//
// Solidity: function getTotalMasterSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiSession) GetTotalMasterSnapshotterCount() (*big.Int, error) {
	return _ContractApi.Contract.GetTotalMasterSnapshotterCount(&_ContractApi.CallOpts)
}

// GetTotalMasterSnapshotterCount is a free data retrieval call binding the contract method 0xd551672b.
//
// Solidity: function getTotalMasterSnapshotterCount() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) GetTotalMasterSnapshotterCount() (*big.Int, error) {
	return _ContractApi.Contract.GetTotalMasterSnapshotterCount(&_ContractApi.CallOpts)
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

// LastFinalizedSnapshot is a free data retrieval call binding the contract method 0x4ea16b0a.
//
// Solidity: function lastFinalizedSnapshot(string ) view returns(uint256)
func (_ContractApi *ContractApiCaller) LastFinalizedSnapshot(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "lastFinalizedSnapshot", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastFinalizedSnapshot is a free data retrieval call binding the contract method 0x4ea16b0a.
//
// Solidity: function lastFinalizedSnapshot(string ) view returns(uint256)
func (_ContractApi *ContractApiSession) LastFinalizedSnapshot(arg0 string) (*big.Int, error) {
	return _ContractApi.Contract.LastFinalizedSnapshot(&_ContractApi.CallOpts, arg0)
}

// LastFinalizedSnapshot is a free data retrieval call binding the contract method 0x4ea16b0a.
//
// Solidity: function lastFinalizedSnapshot(string ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) LastFinalizedSnapshot(arg0 string) (*big.Int, error) {
	return _ContractApi.Contract.LastFinalizedSnapshot(&_ContractApi.CallOpts, arg0)
}

// LastSnapshotterAddressUpdate is a free data retrieval call binding the contract method 0x25d6ad01.
//
// Solidity: function lastSnapshotterAddressUpdate(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) LastSnapshotterAddressUpdate(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "lastSnapshotterAddressUpdate", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastSnapshotterAddressUpdate is a free data retrieval call binding the contract method 0x25d6ad01.
//
// Solidity: function lastSnapshotterAddressUpdate(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) LastSnapshotterAddressUpdate(arg0 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.LastSnapshotterAddressUpdate(&_ContractApi.CallOpts, arg0)
}

// LastSnapshotterAddressUpdate is a free data retrieval call binding the contract method 0x25d6ad01.
//
// Solidity: function lastSnapshotterAddressUpdate(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) LastSnapshotterAddressUpdate(arg0 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.LastSnapshotterAddressUpdate(&_ContractApi.CallOpts, arg0)
}

// MasterSnapshotters is a free data retrieval call binding the contract method 0x34b739d9.
//
// Solidity: function masterSnapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiCaller) MasterSnapshotters(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "masterSnapshotters", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MasterSnapshotters is a free data retrieval call binding the contract method 0x34b739d9.
//
// Solidity: function masterSnapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiSession) MasterSnapshotters(arg0 common.Address) (bool, error) {
	return _ContractApi.Contract.MasterSnapshotters(&_ContractApi.CallOpts, arg0)
}

// MasterSnapshotters is a free data retrieval call binding the contract method 0x34b739d9.
//
// Solidity: function masterSnapshotters(address ) view returns(bool)
func (_ContractApi *ContractApiCallerSession) MasterSnapshotters(arg0 common.Address) (bool, error) {
	return _ContractApi.Contract.MasterSnapshotters(&_ContractApi.CallOpts, arg0)
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

// RewardBasePoints is a free data retrieval call binding the contract method 0x7e9834d7.
//
// Solidity: function rewardBasePoints() view returns(uint256)
func (_ContractApi *ContractApiCaller) RewardBasePoints(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "rewardBasePoints")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardBasePoints is a free data retrieval call binding the contract method 0x7e9834d7.
//
// Solidity: function rewardBasePoints() view returns(uint256)
func (_ContractApi *ContractApiSession) RewardBasePoints() (*big.Int, error) {
	return _ContractApi.Contract.RewardBasePoints(&_ContractApi.CallOpts)
}

// RewardBasePoints is a free data retrieval call binding the contract method 0x7e9834d7.
//
// Solidity: function rewardBasePoints() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) RewardBasePoints() (*big.Int, error) {
	return _ContractApi.Contract.RewardBasePoints(&_ContractApi.CallOpts)
}

// RewardsEnabled is a free data retrieval call binding the contract method 0x1dafe16b.
//
// Solidity: function rewardsEnabled() view returns(bool)
func (_ContractApi *ContractApiCaller) RewardsEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "rewardsEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RewardsEnabled is a free data retrieval call binding the contract method 0x1dafe16b.
//
// Solidity: function rewardsEnabled() view returns(bool)
func (_ContractApi *ContractApiSession) RewardsEnabled() (bool, error) {
	return _ContractApi.Contract.RewardsEnabled(&_ContractApi.CallOpts)
}

// RewardsEnabled is a free data retrieval call binding the contract method 0x1dafe16b.
//
// Solidity: function rewardsEnabled() view returns(bool)
func (_ContractApi *ContractApiCallerSession) RewardsEnabled() (bool, error) {
	return _ContractApi.Contract.RewardsEnabled(&_ContractApi.CallOpts)
}

// SlotCounter is a free data retrieval call binding the contract method 0xe59a4105.
//
// Solidity: function slotCounter() view returns(uint256)
func (_ContractApi *ContractApiCaller) SlotCounter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "slotCounter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotCounter is a free data retrieval call binding the contract method 0xe59a4105.
//
// Solidity: function slotCounter() view returns(uint256)
func (_ContractApi *ContractApiSession) SlotCounter() (*big.Int, error) {
	return _ContractApi.Contract.SlotCounter(&_ContractApi.CallOpts)
}

// SlotCounter is a free data retrieval call binding the contract method 0xe59a4105.
//
// Solidity: function slotCounter() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SlotCounter() (*big.Int, error) {
	return _ContractApi.Contract.SlotCounter(&_ContractApi.CallOpts)
}

// SlotRewardPoints is a free data retrieval call binding the contract method 0x486429e7.
//
// Solidity: function slotRewardPoints(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) SlotRewardPoints(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "slotRewardPoints", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotRewardPoints is a free data retrieval call binding the contract method 0x486429e7.
//
// Solidity: function slotRewardPoints(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) SlotRewardPoints(arg0 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.SlotRewardPoints(&_ContractApi.CallOpts, arg0)
}

// SlotRewardPoints is a free data retrieval call binding the contract method 0x486429e7.
//
// Solidity: function slotRewardPoints(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SlotRewardPoints(arg0 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.SlotRewardPoints(&_ContractApi.CallOpts, arg0)
}

// SlotSnapshotterMapping is a free data retrieval call binding the contract method 0x948a463e.
//
// Solidity: function slotSnapshotterMapping(uint256 ) view returns(address)
func (_ContractApi *ContractApiCaller) SlotSnapshotterMapping(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "slotSnapshotterMapping", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SlotSnapshotterMapping is a free data retrieval call binding the contract method 0x948a463e.
//
// Solidity: function slotSnapshotterMapping(uint256 ) view returns(address)
func (_ContractApi *ContractApiSession) SlotSnapshotterMapping(arg0 *big.Int) (common.Address, error) {
	return _ContractApi.Contract.SlotSnapshotterMapping(&_ContractApi.CallOpts, arg0)
}

// SlotSnapshotterMapping is a free data retrieval call binding the contract method 0x948a463e.
//
// Solidity: function slotSnapshotterMapping(uint256 ) view returns(address)
func (_ContractApi *ContractApiCallerSession) SlotSnapshotterMapping(arg0 *big.Int) (common.Address, error) {
	return _ContractApi.Contract.SlotSnapshotterMapping(&_ContractApi.CallOpts, arg0)
}

// SlotStreakCounter is a free data retrieval call binding the contract method 0x7c5f1557.
//
// Solidity: function slotStreakCounter(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) SlotStreakCounter(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "slotStreakCounter", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotStreakCounter is a free data retrieval call binding the contract method 0x7c5f1557.
//
// Solidity: function slotStreakCounter(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) SlotStreakCounter(arg0 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.SlotStreakCounter(&_ContractApi.CallOpts, arg0)
}

// SlotStreakCounter is a free data retrieval call binding the contract method 0x7c5f1557.
//
// Solidity: function slotStreakCounter(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SlotStreakCounter(arg0 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.SlotStreakCounter(&_ContractApi.CallOpts, arg0)
}

// SlotSubmissionCount is a free data retrieval call binding the contract method 0x9dbc5064.
//
// Solidity: function slotSubmissionCount(uint256 , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) SlotSubmissionCount(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "slotSubmissionCount", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotSubmissionCount is a free data retrieval call binding the contract method 0x9dbc5064.
//
// Solidity: function slotSubmissionCount(uint256 , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) SlotSubmissionCount(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.SlotSubmissionCount(&_ContractApi.CallOpts, arg0, arg1)
}

// SlotSubmissionCount is a free data retrieval call binding the contract method 0x9dbc5064.
//
// Solidity: function slotSubmissionCount(uint256 , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) SlotSubmissionCount(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.SlotSubmissionCount(&_ContractApi.CallOpts, arg0, arg1)
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

// SnapshotsReceivedSlot is a free data retrieval call binding the contract method 0x37d080f0.
//
// Solidity: function snapshotsReceivedSlot(string , uint256 , uint256 ) view returns(bool)
func (_ContractApi *ContractApiCaller) SnapshotsReceivedSlot(opts *bind.CallOpts, arg0 string, arg1 *big.Int, arg2 *big.Int) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "snapshotsReceivedSlot", arg0, arg1, arg2)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SnapshotsReceivedSlot is a free data retrieval call binding the contract method 0x37d080f0.
//
// Solidity: function snapshotsReceivedSlot(string , uint256 , uint256 ) view returns(bool)
func (_ContractApi *ContractApiSession) SnapshotsReceivedSlot(arg0 string, arg1 *big.Int, arg2 *big.Int) (bool, error) {
	return _ContractApi.Contract.SnapshotsReceivedSlot(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceivedSlot is a free data retrieval call binding the contract method 0x37d080f0.
//
// Solidity: function snapshotsReceivedSlot(string , uint256 , uint256 ) view returns(bool)
func (_ContractApi *ContractApiCallerSession) SnapshotsReceivedSlot(arg0 string, arg1 *big.Int, arg2 *big.Int) (bool, error) {
	return _ContractApi.Contract.SnapshotsReceivedSlot(&_ContractApi.CallOpts, arg0, arg1, arg2)
}

// StreakBonusPoints is a free data retrieval call binding the contract method 0x36f71682.
//
// Solidity: function streakBonusPoints() view returns(uint256)
func (_ContractApi *ContractApiCaller) StreakBonusPoints(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "streakBonusPoints")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StreakBonusPoints is a free data retrieval call binding the contract method 0x36f71682.
//
// Solidity: function streakBonusPoints() view returns(uint256)
func (_ContractApi *ContractApiSession) StreakBonusPoints() (*big.Int, error) {
	return _ContractApi.Contract.StreakBonusPoints(&_ContractApi.CallOpts)
}

// StreakBonusPoints is a free data retrieval call binding the contract method 0x36f71682.
//
// Solidity: function streakBonusPoints() view returns(uint256)
func (_ContractApi *ContractApiCallerSession) StreakBonusPoints() (*big.Int, error) {
	return _ContractApi.Contract.StreakBonusPoints(&_ContractApi.CallOpts)
}

// TimeSlotCheck is a free data retrieval call binding the contract method 0x1a0adc24.
//
// Solidity: function timeSlotCheck() view returns(bool)
func (_ContractApi *ContractApiCaller) TimeSlotCheck(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "timeSlotCheck")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TimeSlotCheck is a free data retrieval call binding the contract method 0x1a0adc24.
//
// Solidity: function timeSlotCheck() view returns(bool)
func (_ContractApi *ContractApiSession) TimeSlotCheck() (bool, error) {
	return _ContractApi.Contract.TimeSlotCheck(&_ContractApi.CallOpts)
}

// TimeSlotCheck is a free data retrieval call binding the contract method 0x1a0adc24.
//
// Solidity: function timeSlotCheck() view returns(bool)
func (_ContractApi *ContractApiCallerSession) TimeSlotCheck() (bool, error) {
	return _ContractApi.Contract.TimeSlotCheck(&_ContractApi.CallOpts)
}

// TimeSlotPreference is a free data retrieval call binding the contract method 0x1cdd886b.
//
// Solidity: function timeSlotPreference(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) TimeSlotPreference(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "timeSlotPreference", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TimeSlotPreference is a free data retrieval call binding the contract method 0x1cdd886b.
//
// Solidity: function timeSlotPreference(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) TimeSlotPreference(arg0 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.TimeSlotPreference(&_ContractApi.CallOpts, arg0)
}

// TimeSlotPreference is a free data retrieval call binding the contract method 0x1cdd886b.
//
// Solidity: function timeSlotPreference(uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) TimeSlotPreference(arg0 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.TimeSlotPreference(&_ContractApi.CallOpts, arg0)
}

// TotalSnapshotsReceived is a free data retrieval call binding the contract method 0x797bbd58.
//
// Solidity: function totalSnapshotsReceived(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCaller) TotalSnapshotsReceived(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "totalSnapshotsReceived", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSnapshotsReceived is a free data retrieval call binding the contract method 0x797bbd58.
//
// Solidity: function totalSnapshotsReceived(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiSession) TotalSnapshotsReceived(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.TotalSnapshotsReceived(&_ContractApi.CallOpts, arg0, arg1)
}

// TotalSnapshotsReceived is a free data retrieval call binding the contract method 0x797bbd58.
//
// Solidity: function totalSnapshotsReceived(string , uint256 ) view returns(uint256)
func (_ContractApi *ContractApiCallerSession) TotalSnapshotsReceived(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ContractApi.Contract.TotalSnapshotsReceived(&_ContractApi.CallOpts, arg0, arg1)
}

// Verify is a free data retrieval call binding the contract method 0x58939f83.
//
// Solidity: function verify(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiCaller) Verify(opts *bind.CallOpts, slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	var out []interface{}
	err := _ContractApi.contract.Call(opts, &out, "verify", slotId, snapshotCid, epochId, projectId, request, signature, signer)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Verify is a free data retrieval call binding the contract method 0x58939f83.
//
// Solidity: function verify(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiSession) Verify(slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	return _ContractApi.Contract.Verify(&_ContractApi.CallOpts, slotId, snapshotCid, epochId, projectId, request, signature, signer)
}

// Verify is a free data retrieval call binding the contract method 0x58939f83.
//
// Solidity: function verify(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ContractApi *ContractApiCallerSession) Verify(slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	return _ContractApi.Contract.Verify(&_ContractApi.CallOpts, slotId, snapshotCid, epochId, projectId, request, signature, signer)
}

// AssignSnapshotterToSlotWithTimeSlot is a paid mutator transaction binding the contract method 0x7341fd46.
//
// Solidity: function assignSnapshotterToSlotWithTimeSlot(uint256 slotId, address snapshotterAddress, uint256 timeSlot) returns()
func (_ContractApi *ContractApiTransactor) AssignSnapshotterToSlotWithTimeSlot(opts *bind.TransactOpts, slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "assignSnapshotterToSlotWithTimeSlot", slotId, snapshotterAddress, timeSlot)
}

// AssignSnapshotterToSlotWithTimeSlot is a paid mutator transaction binding the contract method 0x7341fd46.
//
// Solidity: function assignSnapshotterToSlotWithTimeSlot(uint256 slotId, address snapshotterAddress, uint256 timeSlot) returns()
func (_ContractApi *ContractApiSession) AssignSnapshotterToSlotWithTimeSlot(slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.AssignSnapshotterToSlotWithTimeSlot(&_ContractApi.TransactOpts, slotId, snapshotterAddress, timeSlot)
}

// AssignSnapshotterToSlotWithTimeSlot is a paid mutator transaction binding the contract method 0x7341fd46.
//
// Solidity: function assignSnapshotterToSlotWithTimeSlot(uint256 slotId, address snapshotterAddress, uint256 timeSlot) returns()
func (_ContractApi *ContractApiTransactorSession) AssignSnapshotterToSlotWithTimeSlot(slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.AssignSnapshotterToSlotWithTimeSlot(&_ContractApi.TransactOpts, slotId, snapshotterAddress, timeSlot)
}

// AssignSnapshotterToSlotsWithTimeSlot is a paid mutator transaction binding the contract method 0xcd6d56bd.
//
// Solidity: function assignSnapshotterToSlotsWithTimeSlot(uint256[] slotIds, address[] snapshotterAddresses, uint256[] timeSlots) returns()
func (_ContractApi *ContractApiTransactor) AssignSnapshotterToSlotsWithTimeSlot(opts *bind.TransactOpts, slotIds []*big.Int, snapshotterAddresses []common.Address, timeSlots []*big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "assignSnapshotterToSlotsWithTimeSlot", slotIds, snapshotterAddresses, timeSlots)
}

// AssignSnapshotterToSlotsWithTimeSlot is a paid mutator transaction binding the contract method 0xcd6d56bd.
//
// Solidity: function assignSnapshotterToSlotsWithTimeSlot(uint256[] slotIds, address[] snapshotterAddresses, uint256[] timeSlots) returns()
func (_ContractApi *ContractApiSession) AssignSnapshotterToSlotsWithTimeSlot(slotIds []*big.Int, snapshotterAddresses []common.Address, timeSlots []*big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.AssignSnapshotterToSlotsWithTimeSlot(&_ContractApi.TransactOpts, slotIds, snapshotterAddresses, timeSlots)
}

// AssignSnapshotterToSlotsWithTimeSlot is a paid mutator transaction binding the contract method 0xcd6d56bd.
//
// Solidity: function assignSnapshotterToSlotsWithTimeSlot(uint256[] slotIds, address[] snapshotterAddresses, uint256[] timeSlots) returns()
func (_ContractApi *ContractApiTransactorSession) AssignSnapshotterToSlotsWithTimeSlot(slotIds []*big.Int, snapshotterAddresses []common.Address, timeSlots []*big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.AssignSnapshotterToSlotsWithTimeSlot(&_ContractApi.TransactOpts, slotIds, snapshotterAddresses, timeSlots)
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

// ForceSkipEpoch is a paid mutator transaction binding the contract method 0xf537a3e2.
//
// Solidity: function forceSkipEpoch(uint256 begin, uint256 end) returns()
func (_ContractApi *ContractApiTransactor) ForceSkipEpoch(opts *bind.TransactOpts, begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "forceSkipEpoch", begin, end)
}

// ForceSkipEpoch is a paid mutator transaction binding the contract method 0xf537a3e2.
//
// Solidity: function forceSkipEpoch(uint256 begin, uint256 end) returns()
func (_ContractApi *ContractApiSession) ForceSkipEpoch(begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceSkipEpoch(&_ContractApi.TransactOpts, begin, end)
}

// ForceSkipEpoch is a paid mutator transaction binding the contract method 0xf537a3e2.
//
// Solidity: function forceSkipEpoch(uint256 begin, uint256 end) returns()
func (_ContractApi *ContractApiTransactorSession) ForceSkipEpoch(begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.ForceSkipEpoch(&_ContractApi.TransactOpts, begin, end)
}

// ForceStartDay is a paid mutator transaction binding the contract method 0xcd529926.
//
// Solidity: function forceStartDay() returns()
func (_ContractApi *ContractApiTransactor) ForceStartDay(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "forceStartDay")
}

// ForceStartDay is a paid mutator transaction binding the contract method 0xcd529926.
//
// Solidity: function forceStartDay() returns()
func (_ContractApi *ContractApiSession) ForceStartDay() (*types.Transaction, error) {
	return _ContractApi.Contract.ForceStartDay(&_ContractApi.TransactOpts)
}

// ForceStartDay is a paid mutator transaction binding the contract method 0xcd529926.
//
// Solidity: function forceStartDay() returns()
func (_ContractApi *ContractApiTransactorSession) ForceStartDay() (*types.Transaction, error) {
	return _ContractApi.Contract.ForceStartDay(&_ContractApi.TransactOpts)
}

// LoadCurrentDay is a paid mutator transaction binding the contract method 0x82cdfd43.
//
// Solidity: function loadCurrentDay(uint256 _dayCounter) returns()
func (_ContractApi *ContractApiTransactor) LoadCurrentDay(opts *bind.TransactOpts, _dayCounter *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "loadCurrentDay", _dayCounter)
}

// LoadCurrentDay is a paid mutator transaction binding the contract method 0x82cdfd43.
//
// Solidity: function loadCurrentDay(uint256 _dayCounter) returns()
func (_ContractApi *ContractApiSession) LoadCurrentDay(_dayCounter *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.LoadCurrentDay(&_ContractApi.TransactOpts, _dayCounter)
}

// LoadCurrentDay is a paid mutator transaction binding the contract method 0x82cdfd43.
//
// Solidity: function loadCurrentDay(uint256 _dayCounter) returns()
func (_ContractApi *ContractApiTransactorSession) LoadCurrentDay(_dayCounter *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.LoadCurrentDay(&_ContractApi.TransactOpts, _dayCounter)
}

// LoadSlotInfo is a paid mutator transaction binding the contract method 0xa0c6d7f3.
//
// Solidity: function loadSlotInfo(uint256 slotId, address snapshotterAddress, uint256 timeSlot, uint256 rewardPoints, uint256 currentStreak, uint256 currentDaySnapshotCount) returns()
func (_ContractApi *ContractApiTransactor) LoadSlotInfo(opts *bind.TransactOpts, slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int, rewardPoints *big.Int, currentStreak *big.Int, currentDaySnapshotCount *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "loadSlotInfo", slotId, snapshotterAddress, timeSlot, rewardPoints, currentStreak, currentDaySnapshotCount)
}

// LoadSlotInfo is a paid mutator transaction binding the contract method 0xa0c6d7f3.
//
// Solidity: function loadSlotInfo(uint256 slotId, address snapshotterAddress, uint256 timeSlot, uint256 rewardPoints, uint256 currentStreak, uint256 currentDaySnapshotCount) returns()
func (_ContractApi *ContractApiSession) LoadSlotInfo(slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int, rewardPoints *big.Int, currentStreak *big.Int, currentDaySnapshotCount *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.LoadSlotInfo(&_ContractApi.TransactOpts, slotId, snapshotterAddress, timeSlot, rewardPoints, currentStreak, currentDaySnapshotCount)
}

// LoadSlotInfo is a paid mutator transaction binding the contract method 0xa0c6d7f3.
//
// Solidity: function loadSlotInfo(uint256 slotId, address snapshotterAddress, uint256 timeSlot, uint256 rewardPoints, uint256 currentStreak, uint256 currentDaySnapshotCount) returns()
func (_ContractApi *ContractApiTransactorSession) LoadSlotInfo(slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int, rewardPoints *big.Int, currentStreak *big.Int, currentDaySnapshotCount *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.LoadSlotInfo(&_ContractApi.TransactOpts, slotId, snapshotterAddress, timeSlot, rewardPoints, currentStreak, currentDaySnapshotCount)
}

// LoadSlotSubmissions is a paid mutator transaction binding the contract method 0xbb8ab44b.
//
// Solidity: function loadSlotSubmissions(uint256 slotId, uint256 dayId, uint256 snapshotCount) returns()
func (_ContractApi *ContractApiTransactor) LoadSlotSubmissions(opts *bind.TransactOpts, slotId *big.Int, dayId *big.Int, snapshotCount *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "loadSlotSubmissions", slotId, dayId, snapshotCount)
}

// LoadSlotSubmissions is a paid mutator transaction binding the contract method 0xbb8ab44b.
//
// Solidity: function loadSlotSubmissions(uint256 slotId, uint256 dayId, uint256 snapshotCount) returns()
func (_ContractApi *ContractApiSession) LoadSlotSubmissions(slotId *big.Int, dayId *big.Int, snapshotCount *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.LoadSlotSubmissions(&_ContractApi.TransactOpts, slotId, dayId, snapshotCount)
}

// LoadSlotSubmissions is a paid mutator transaction binding the contract method 0xbb8ab44b.
//
// Solidity: function loadSlotSubmissions(uint256 slotId, uint256 dayId, uint256 snapshotCount) returns()
func (_ContractApi *ContractApiTransactorSession) LoadSlotSubmissions(slotId *big.Int, dayId *big.Int, snapshotCount *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.LoadSlotSubmissions(&_ContractApi.TransactOpts, slotId, dayId, snapshotCount)
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

// SubmitSnapshot is a paid mutator transaction binding the contract method 0x0fbe203a.
//
// Solidity: function submitSnapshot(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactor) SubmitSnapshot(opts *bind.TransactOpts, slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "submitSnapshot", slotId, snapshotCid, epochId, projectId, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0x0fbe203a.
//
// Solidity: function submitSnapshot(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature) returns()
func (_ContractApi *ContractApiSession) SubmitSnapshot(slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitSnapshot(&_ContractApi.TransactOpts, slotId, snapshotCid, epochId, projectId, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0x0fbe203a.
//
// Solidity: function submitSnapshot(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature) returns()
func (_ContractApi *ContractApiTransactorSession) SubmitSnapshot(slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ContractApi.Contract.SubmitSnapshot(&_ContractApi.TransactOpts, slotId, snapshotCid, epochId, projectId, request, signature)
}

// ToggleRewards is a paid mutator transaction binding the contract method 0x95268408.
//
// Solidity: function toggleRewards() returns()
func (_ContractApi *ContractApiTransactor) ToggleRewards(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "toggleRewards")
}

// ToggleRewards is a paid mutator transaction binding the contract method 0x95268408.
//
// Solidity: function toggleRewards() returns()
func (_ContractApi *ContractApiSession) ToggleRewards() (*types.Transaction, error) {
	return _ContractApi.Contract.ToggleRewards(&_ContractApi.TransactOpts)
}

// ToggleRewards is a paid mutator transaction binding the contract method 0x95268408.
//
// Solidity: function toggleRewards() returns()
func (_ContractApi *ContractApiTransactorSession) ToggleRewards() (*types.Transaction, error) {
	return _ContractApi.Contract.ToggleRewards(&_ContractApi.TransactOpts)
}

// ToggleTimeSlotCheck is a paid mutator transaction binding the contract method 0x75f2d951.
//
// Solidity: function toggleTimeSlotCheck() returns()
func (_ContractApi *ContractApiTransactor) ToggleTimeSlotCheck(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "toggleTimeSlotCheck")
}

// ToggleTimeSlotCheck is a paid mutator transaction binding the contract method 0x75f2d951.
//
// Solidity: function toggleTimeSlotCheck() returns()
func (_ContractApi *ContractApiSession) ToggleTimeSlotCheck() (*types.Transaction, error) {
	return _ContractApi.Contract.ToggleTimeSlotCheck(&_ContractApi.TransactOpts)
}

// ToggleTimeSlotCheck is a paid mutator transaction binding the contract method 0x75f2d951.
//
// Solidity: function toggleTimeSlotCheck() returns()
func (_ContractApi *ContractApiTransactorSession) ToggleTimeSlotCheck() (*types.Transaction, error) {
	return _ContractApi.Contract.ToggleTimeSlotCheck(&_ContractApi.TransactOpts)
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

// UpdateAllowedProjectType is a paid mutator transaction binding the contract method 0x562209f3.
//
// Solidity: function updateAllowedProjectType(string _projectType, bool _status) returns()
func (_ContractApi *ContractApiTransactor) UpdateAllowedProjectType(opts *bind.TransactOpts, _projectType string, _status bool) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateAllowedProjectType", _projectType, _status)
}

// UpdateAllowedProjectType is a paid mutator transaction binding the contract method 0x562209f3.
//
// Solidity: function updateAllowedProjectType(string _projectType, bool _status) returns()
func (_ContractApi *ContractApiSession) UpdateAllowedProjectType(_projectType string, _status bool) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateAllowedProjectType(&_ContractApi.TransactOpts, _projectType, _status)
}

// UpdateAllowedProjectType is a paid mutator transaction binding the contract method 0x562209f3.
//
// Solidity: function updateAllowedProjectType(string _projectType, bool _status) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateAllowedProjectType(_projectType string, _status bool) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateAllowedProjectType(&_ContractApi.TransactOpts, _projectType, _status)
}

// UpdateDailySnapshotQuota is a paid mutator transaction binding the contract method 0x6d26e630.
//
// Solidity: function updateDailySnapshotQuota(uint256 dailySnapshotQuota) returns()
func (_ContractApi *ContractApiTransactor) UpdateDailySnapshotQuota(opts *bind.TransactOpts, dailySnapshotQuota *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateDailySnapshotQuota", dailySnapshotQuota)
}

// UpdateDailySnapshotQuota is a paid mutator transaction binding the contract method 0x6d26e630.
//
// Solidity: function updateDailySnapshotQuota(uint256 dailySnapshotQuota) returns()
func (_ContractApi *ContractApiSession) UpdateDailySnapshotQuota(dailySnapshotQuota *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateDailySnapshotQuota(&_ContractApi.TransactOpts, dailySnapshotQuota)
}

// UpdateDailySnapshotQuota is a paid mutator transaction binding the contract method 0x6d26e630.
//
// Solidity: function updateDailySnapshotQuota(uint256 dailySnapshotQuota) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateDailySnapshotQuota(dailySnapshotQuota *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateDailySnapshotQuota(&_ContractApi.TransactOpts, dailySnapshotQuota)
}

// UpdateMasterSnapshotters is a paid mutator transaction binding the contract method 0x8f9eacf3.
//
// Solidity: function updateMasterSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ContractApi *ContractApiTransactor) UpdateMasterSnapshotters(opts *bind.TransactOpts, _snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updateMasterSnapshotters", _snapshotters, _status)
}

// UpdateMasterSnapshotters is a paid mutator transaction binding the contract method 0x8f9eacf3.
//
// Solidity: function updateMasterSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ContractApi *ContractApiSession) UpdateMasterSnapshotters(_snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateMasterSnapshotters(&_ContractApi.TransactOpts, _snapshotters, _status)
}

// UpdateMasterSnapshotters is a paid mutator transaction binding the contract method 0x8f9eacf3.
//
// Solidity: function updateMasterSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ContractApi *ContractApiTransactorSession) UpdateMasterSnapshotters(_snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdateMasterSnapshotters(&_ContractApi.TransactOpts, _snapshotters, _status)
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

// UpdaterewardBasePoints is a paid mutator transaction binding the contract method 0x801939d7.
//
// Solidity: function updaterewardBasePoints(uint256 newrewardBasePoints) returns()
func (_ContractApi *ContractApiTransactor) UpdaterewardBasePoints(opts *bind.TransactOpts, newrewardBasePoints *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updaterewardBasePoints", newrewardBasePoints)
}

// UpdaterewardBasePoints is a paid mutator transaction binding the contract method 0x801939d7.
//
// Solidity: function updaterewardBasePoints(uint256 newrewardBasePoints) returns()
func (_ContractApi *ContractApiSession) UpdaterewardBasePoints(newrewardBasePoints *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdaterewardBasePoints(&_ContractApi.TransactOpts, newrewardBasePoints)
}

// UpdaterewardBasePoints is a paid mutator transaction binding the contract method 0x801939d7.
//
// Solidity: function updaterewardBasePoints(uint256 newrewardBasePoints) returns()
func (_ContractApi *ContractApiTransactorSession) UpdaterewardBasePoints(newrewardBasePoints *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdaterewardBasePoints(&_ContractApi.TransactOpts, newrewardBasePoints)
}

// UpdatestreakBonusPoints is a paid mutator transaction binding the contract method 0xa8939fa5.
//
// Solidity: function updatestreakBonusPoints(uint256 newstreakBonusPoints) returns()
func (_ContractApi *ContractApiTransactor) UpdatestreakBonusPoints(opts *bind.TransactOpts, newstreakBonusPoints *big.Int) (*types.Transaction, error) {
	return _ContractApi.contract.Transact(opts, "updatestreakBonusPoints", newstreakBonusPoints)
}

// UpdatestreakBonusPoints is a paid mutator transaction binding the contract method 0xa8939fa5.
//
// Solidity: function updatestreakBonusPoints(uint256 newstreakBonusPoints) returns()
func (_ContractApi *ContractApiSession) UpdatestreakBonusPoints(newstreakBonusPoints *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdatestreakBonusPoints(&_ContractApi.TransactOpts, newstreakBonusPoints)
}

// UpdatestreakBonusPoints is a paid mutator transaction binding the contract method 0xa8939fa5.
//
// Solidity: function updatestreakBonusPoints(uint256 newstreakBonusPoints) returns()
func (_ContractApi *ContractApiTransactorSession) UpdatestreakBonusPoints(newstreakBonusPoints *big.Int) (*types.Transaction, error) {
	return _ContractApi.Contract.UpdatestreakBonusPoints(&_ContractApi.TransactOpts, newstreakBonusPoints)
}

// ContractApiDailyTaskCompletedEventIterator is returned from FilterDailyTaskCompletedEvent and is used to iterate over the raw logs and unpacked data for DailyTaskCompletedEvent events raised by the ContractApi contract.
type ContractApiDailyTaskCompletedEventIterator struct {
	Event *ContractApiDailyTaskCompletedEvent // Event containing the contract specifics and raw log

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
func (it *ContractApiDailyTaskCompletedEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiDailyTaskCompletedEvent)
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
		it.Event = new(ContractApiDailyTaskCompletedEvent)
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
func (it *ContractApiDailyTaskCompletedEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiDailyTaskCompletedEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiDailyTaskCompletedEvent represents a DailyTaskCompletedEvent event raised by the ContractApi contract.
type ContractApiDailyTaskCompletedEvent struct {
	SnapshotterAddress common.Address
	SlotId             *big.Int
	DayId              *big.Int
	Timestamp          *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterDailyTaskCompletedEvent is a free log retrieval operation binding the contract event 0x34c900c4105cef3bd58c4b7d2b6fe54f1f64845d5bd5ed2e2e92b52aed2d58ae.
//
// Solidity: event DailyTaskCompletedEvent(address snapshotterAddress, uint256 slotId, uint256 dayId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) FilterDailyTaskCompletedEvent(opts *bind.FilterOpts) (*ContractApiDailyTaskCompletedEventIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "DailyTaskCompletedEvent")
	if err != nil {
		return nil, err
	}
	return &ContractApiDailyTaskCompletedEventIterator{contract: _ContractApi.contract, event: "DailyTaskCompletedEvent", logs: logs, sub: sub}, nil
}

// WatchDailyTaskCompletedEvent is a free log subscription operation binding the contract event 0x34c900c4105cef3bd58c4b7d2b6fe54f1f64845d5bd5ed2e2e92b52aed2d58ae.
//
// Solidity: event DailyTaskCompletedEvent(address snapshotterAddress, uint256 slotId, uint256 dayId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) WatchDailyTaskCompletedEvent(opts *bind.WatchOpts, sink chan<- *ContractApiDailyTaskCompletedEvent) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "DailyTaskCompletedEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiDailyTaskCompletedEvent)
				if err := _ContractApi.contract.UnpackLog(event, "DailyTaskCompletedEvent", log); err != nil {
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

// ParseDailyTaskCompletedEvent is a log parse operation binding the contract event 0x34c900c4105cef3bd58c4b7d2b6fe54f1f64845d5bd5ed2e2e92b52aed2d58ae.
//
// Solidity: event DailyTaskCompletedEvent(address snapshotterAddress, uint256 slotId, uint256 dayId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) ParseDailyTaskCompletedEvent(log types.Log) (*ContractApiDailyTaskCompletedEvent, error) {
	event := new(ContractApiDailyTaskCompletedEvent)
	if err := _ContractApi.contract.UnpackLog(event, "DailyTaskCompletedEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiDayStartedEventIterator is returned from FilterDayStartedEvent and is used to iterate over the raw logs and unpacked data for DayStartedEvent events raised by the ContractApi contract.
type ContractApiDayStartedEventIterator struct {
	Event *ContractApiDayStartedEvent // Event containing the contract specifics and raw log

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
func (it *ContractApiDayStartedEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiDayStartedEvent)
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
		it.Event = new(ContractApiDayStartedEvent)
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
func (it *ContractApiDayStartedEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiDayStartedEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiDayStartedEvent represents a DayStartedEvent event raised by the ContractApi contract.
type ContractApiDayStartedEvent struct {
	DayId     *big.Int
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDayStartedEvent is a free log retrieval operation binding the contract event 0xf391963fbbcec4cbb1f4a6c915c531364db26c103d31434223c3bddb703c94fe.
//
// Solidity: event DayStartedEvent(uint256 dayId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) FilterDayStartedEvent(opts *bind.FilterOpts) (*ContractApiDayStartedEventIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "DayStartedEvent")
	if err != nil {
		return nil, err
	}
	return &ContractApiDayStartedEventIterator{contract: _ContractApi.contract, event: "DayStartedEvent", logs: logs, sub: sub}, nil
}

// WatchDayStartedEvent is a free log subscription operation binding the contract event 0xf391963fbbcec4cbb1f4a6c915c531364db26c103d31434223c3bddb703c94fe.
//
// Solidity: event DayStartedEvent(uint256 dayId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) WatchDayStartedEvent(opts *bind.WatchOpts, sink chan<- *ContractApiDayStartedEvent) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "DayStartedEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiDayStartedEvent)
				if err := _ContractApi.contract.UnpackLog(event, "DayStartedEvent", log); err != nil {
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

// ParseDayStartedEvent is a log parse operation binding the contract event 0xf391963fbbcec4cbb1f4a6c915c531364db26c103d31434223c3bddb703c94fe.
//
// Solidity: event DayStartedEvent(uint256 dayId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) ParseDayStartedEvent(log types.Log) (*ContractApiDayStartedEvent, error) {
	event := new(ContractApiDayStartedEvent)
	if err := _ContractApi.contract.UnpackLog(event, "DayStartedEvent", log); err != nil {
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
	SlotId          *big.Int
	SnapshotCid     string
	EpochId         *big.Int
	ProjectId       string
	Timestamp       *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterDelayedSnapshotSubmitted is a free log retrieval operation binding the contract event 0x8eb09d3de43d012d958db78fcc692d37e9f9c0cc2af1cdd8c58617832af17d31.
//
// Solidity: event DelayedSnapshotSubmitted(address indexed snapshotterAddr, uint256 slotId, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) FilterDelayedSnapshotSubmitted(opts *bind.FilterOpts, snapshotterAddr []common.Address, epochId []*big.Int) (*ContractApiDelayedSnapshotSubmittedIterator, error) {

	var snapshotterAddrRule []interface{}
	for _, snapshotterAddrItem := range snapshotterAddr {
		snapshotterAddrRule = append(snapshotterAddrRule, snapshotterAddrItem)
	}

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "DelayedSnapshotSubmitted", snapshotterAddrRule, epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiDelayedSnapshotSubmittedIterator{contract: _ContractApi.contract, event: "DelayedSnapshotSubmitted", logs: logs, sub: sub}, nil
}

// WatchDelayedSnapshotSubmitted is a free log subscription operation binding the contract event 0x8eb09d3de43d012d958db78fcc692d37e9f9c0cc2af1cdd8c58617832af17d31.
//
// Solidity: event DelayedSnapshotSubmitted(address indexed snapshotterAddr, uint256 slotId, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) WatchDelayedSnapshotSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiDelayedSnapshotSubmitted, snapshotterAddr []common.Address, epochId []*big.Int) (event.Subscription, error) {

	var snapshotterAddrRule []interface{}
	for _, snapshotterAddrItem := range snapshotterAddr {
		snapshotterAddrRule = append(snapshotterAddrRule, snapshotterAddrItem)
	}

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "DelayedSnapshotSubmitted", snapshotterAddrRule, epochIdRule)
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

// ParseDelayedSnapshotSubmitted is a log parse operation binding the contract event 0x8eb09d3de43d012d958db78fcc692d37e9f9c0cc2af1cdd8c58617832af17d31.
//
// Solidity: event DelayedSnapshotSubmitted(address indexed snapshotterAddr, uint256 slotId, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) ParseDelayedSnapshotSubmitted(log types.Log) (*ContractApiDelayedSnapshotSubmitted, error) {
	event := new(ContractApiDelayedSnapshotSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "DelayedSnapshotSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the ContractApi contract.
type ContractApiEIP712DomainChangedIterator struct {
	Event *ContractApiEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *ContractApiEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiEIP712DomainChanged)
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
		it.Event = new(ContractApiEIP712DomainChanged)
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
func (it *ContractApiEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiEIP712DomainChanged represents a EIP712DomainChanged event raised by the ContractApi contract.
type ContractApiEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_ContractApi *ContractApiFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*ContractApiEIP712DomainChangedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &ContractApiEIP712DomainChangedIterator{contract: _ContractApi.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_ContractApi *ContractApiFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *ContractApiEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiEIP712DomainChanged)
				if err := _ContractApi.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_ContractApi *ContractApiFilterer) ParseEIP712DomainChanged(log types.Log) (*ContractApiEIP712DomainChanged, error) {
	event := new(ContractApiEIP712DomainChanged)
	if err := _ContractApi.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ContractApiProjectTypeUpdatedIterator is returned from FilterProjectTypeUpdated and is used to iterate over the raw logs and unpacked data for ProjectTypeUpdated events raised by the ContractApi contract.
type ContractApiProjectTypeUpdatedIterator struct {
	Event *ContractApiProjectTypeUpdated // Event containing the contract specifics and raw log

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
func (it *ContractApiProjectTypeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiProjectTypeUpdated)
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
		it.Event = new(ContractApiProjectTypeUpdated)
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
func (it *ContractApiProjectTypeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiProjectTypeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiProjectTypeUpdated represents a ProjectTypeUpdated event raised by the ContractApi contract.
type ContractApiProjectTypeUpdated struct {
	ProjectType   string
	Allowed       bool
	EnableEpochId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterProjectTypeUpdated is a free log retrieval operation binding the contract event 0xcd1b466e0de4e12ecf0bc286450d3e5b4aa88db70272ed5c11508b16acb35bfc.
//
// Solidity: event ProjectTypeUpdated(string projectType, bool allowed, uint256 enableEpochId)
func (_ContractApi *ContractApiFilterer) FilterProjectTypeUpdated(opts *bind.FilterOpts) (*ContractApiProjectTypeUpdatedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "ProjectTypeUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractApiProjectTypeUpdatedIterator{contract: _ContractApi.contract, event: "ProjectTypeUpdated", logs: logs, sub: sub}, nil
}

// WatchProjectTypeUpdated is a free log subscription operation binding the contract event 0xcd1b466e0de4e12ecf0bc286450d3e5b4aa88db70272ed5c11508b16acb35bfc.
//
// Solidity: event ProjectTypeUpdated(string projectType, bool allowed, uint256 enableEpochId)
func (_ContractApi *ContractApiFilterer) WatchProjectTypeUpdated(opts *bind.WatchOpts, sink chan<- *ContractApiProjectTypeUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "ProjectTypeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiProjectTypeUpdated)
				if err := _ContractApi.contract.UnpackLog(event, "ProjectTypeUpdated", log); err != nil {
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

// ParseProjectTypeUpdated is a log parse operation binding the contract event 0xcd1b466e0de4e12ecf0bc286450d3e5b4aa88db70272ed5c11508b16acb35bfc.
//
// Solidity: event ProjectTypeUpdated(string projectType, bool allowed, uint256 enableEpochId)
func (_ContractApi *ContractApiFilterer) ParseProjectTypeUpdated(log types.Log) (*ContractApiProjectTypeUpdated, error) {
	event := new(ContractApiProjectTypeUpdated)
	if err := _ContractApi.contract.UnpackLog(event, "ProjectTypeUpdated", log); err != nil {
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
	EpochId                *big.Int
	EpochEnd               *big.Int
	ProjectId              string
	SnapshotCid            string
	FinalizedSnapshotCount *big.Int
	TotalReceivedCount     *big.Int
	Timestamp              *big.Int
	Raw                    types.Log // Blockchain specific contextual infos
}

// FilterSnapshotFinalized is a free log retrieval operation binding the contract event 0x029bad7ce411d223096bf56c17c32a3659aa2703de0161497e6284e7fc288c35.
//
// Solidity: event SnapshotFinalized(uint256 indexed epochId, uint256 epochEnd, string projectId, string snapshotCid, uint256 finalizedSnapshotCount, uint256 totalReceivedCount, uint256 timestamp)
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

// WatchSnapshotFinalized is a free log subscription operation binding the contract event 0x029bad7ce411d223096bf56c17c32a3659aa2703de0161497e6284e7fc288c35.
//
// Solidity: event SnapshotFinalized(uint256 indexed epochId, uint256 epochEnd, string projectId, string snapshotCid, uint256 finalizedSnapshotCount, uint256 totalReceivedCount, uint256 timestamp)
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

// ParseSnapshotFinalized is a log parse operation binding the contract event 0x029bad7ce411d223096bf56c17c32a3659aa2703de0161497e6284e7fc288c35.
//
// Solidity: event SnapshotFinalized(uint256 indexed epochId, uint256 epochEnd, string projectId, string snapshotCid, uint256 finalizedSnapshotCount, uint256 totalReceivedCount, uint256 timestamp)
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
	SlotId          *big.Int
	SnapshotCid     string
	EpochId         *big.Int
	ProjectId       string
	Timestamp       *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSnapshotSubmitted is a free log retrieval operation binding the contract event 0x8d9341fa1766ade9f55ddb87e37a11648afefce0f76a389675dd56a5d555b8d3.
//
// Solidity: event SnapshotSubmitted(address indexed snapshotterAddr, uint256 slotId, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) FilterSnapshotSubmitted(opts *bind.FilterOpts, snapshotterAddr []common.Address, epochId []*big.Int) (*ContractApiSnapshotSubmittedIterator, error) {

	var snapshotterAddrRule []interface{}
	for _, snapshotterAddrItem := range snapshotterAddr {
		snapshotterAddrRule = append(snapshotterAddrRule, snapshotterAddrItem)
	}

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "SnapshotSubmitted", snapshotterAddrRule, epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ContractApiSnapshotSubmittedIterator{contract: _ContractApi.contract, event: "SnapshotSubmitted", logs: logs, sub: sub}, nil
}

// WatchSnapshotSubmitted is a free log subscription operation binding the contract event 0x8d9341fa1766ade9f55ddb87e37a11648afefce0f76a389675dd56a5d555b8d3.
//
// Solidity: event SnapshotSubmitted(address indexed snapshotterAddr, uint256 slotId, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) WatchSnapshotSubmitted(opts *bind.WatchOpts, sink chan<- *ContractApiSnapshotSubmitted, snapshotterAddr []common.Address, epochId []*big.Int) (event.Subscription, error) {

	var snapshotterAddrRule []interface{}
	for _, snapshotterAddrItem := range snapshotterAddr {
		snapshotterAddrRule = append(snapshotterAddrRule, snapshotterAddrItem)
	}

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "SnapshotSubmitted", snapshotterAddrRule, epochIdRule)
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

// ParseSnapshotSubmitted is a log parse operation binding the contract event 0x8d9341fa1766ade9f55ddb87e37a11648afefce0f76a389675dd56a5d555b8d3.
//
// Solidity: event SnapshotSubmitted(address indexed snapshotterAddr, uint256 slotId, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ContractApi *ContractApiFilterer) ParseSnapshotSubmitted(log types.Log) (*ContractApiSnapshotSubmitted, error) {
	event := new(ContractApiSnapshotSubmitted)
	if err := _ContractApi.contract.UnpackLog(event, "SnapshotSubmitted", log); err != nil {
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

// ContractApiAllSnapshottersUpdatedIterator is returned from FilterAllSnapshottersUpdated and is used to iterate over the raw logs and unpacked data for AllSnapshottersUpdated events raised by the ContractApi contract.
type ContractApiAllSnapshottersUpdatedIterator struct {
	Event *ContractApiAllSnapshottersUpdated // Event containing the contract specifics and raw log

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
func (it *ContractApiAllSnapshottersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiAllSnapshottersUpdated)
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
		it.Event = new(ContractApiAllSnapshottersUpdated)
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
func (it *ContractApiAllSnapshottersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiAllSnapshottersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiAllSnapshottersUpdated represents a AllSnapshottersUpdated event raised by the ContractApi contract.
type ContractApiAllSnapshottersUpdated struct {
	SnapshotterAddress common.Address
	Allowed            bool
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterAllSnapshottersUpdated is a free log retrieval operation binding the contract event 0x743e47fbcd2e3a64a2ab8f5dcdeb4c17f892c17c9f6ab58d3a5c235953d60058.
//
// Solidity: event allSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) FilterAllSnapshottersUpdated(opts *bind.FilterOpts) (*ContractApiAllSnapshottersUpdatedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "allSnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractApiAllSnapshottersUpdatedIterator{contract: _ContractApi.contract, event: "allSnapshottersUpdated", logs: logs, sub: sub}, nil
}

// WatchAllSnapshottersUpdated is a free log subscription operation binding the contract event 0x743e47fbcd2e3a64a2ab8f5dcdeb4c17f892c17c9f6ab58d3a5c235953d60058.
//
// Solidity: event allSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) WatchAllSnapshottersUpdated(opts *bind.WatchOpts, sink chan<- *ContractApiAllSnapshottersUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "allSnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiAllSnapshottersUpdated)
				if err := _ContractApi.contract.UnpackLog(event, "allSnapshottersUpdated", log); err != nil {
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

// ParseAllSnapshottersUpdated is a log parse operation binding the contract event 0x743e47fbcd2e3a64a2ab8f5dcdeb4c17f892c17c9f6ab58d3a5c235953d60058.
//
// Solidity: event allSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) ParseAllSnapshottersUpdated(log types.Log) (*ContractApiAllSnapshottersUpdated, error) {
	event := new(ContractApiAllSnapshottersUpdated)
	if err := _ContractApi.contract.UnpackLog(event, "allSnapshottersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApiMasterSnapshottersUpdatedIterator is returned from FilterMasterSnapshottersUpdated and is used to iterate over the raw logs and unpacked data for MasterSnapshottersUpdated events raised by the ContractApi contract.
type ContractApiMasterSnapshottersUpdatedIterator struct {
	Event *ContractApiMasterSnapshottersUpdated // Event containing the contract specifics and raw log

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
func (it *ContractApiMasterSnapshottersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApiMasterSnapshottersUpdated)
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
		it.Event = new(ContractApiMasterSnapshottersUpdated)
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
func (it *ContractApiMasterSnapshottersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApiMasterSnapshottersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApiMasterSnapshottersUpdated represents a MasterSnapshottersUpdated event raised by the ContractApi contract.
type ContractApiMasterSnapshottersUpdated struct {
	SnapshotterAddress common.Address
	Allowed            bool
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterMasterSnapshottersUpdated is a free log retrieval operation binding the contract event 0x9c2d4c2b4cf1ca90e31b1448712dae908d3568785b68a411f6617479c2b9913b.
//
// Solidity: event masterSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) FilterMasterSnapshottersUpdated(opts *bind.FilterOpts) (*ContractApiMasterSnapshottersUpdatedIterator, error) {

	logs, sub, err := _ContractApi.contract.FilterLogs(opts, "masterSnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractApiMasterSnapshottersUpdatedIterator{contract: _ContractApi.contract, event: "masterSnapshottersUpdated", logs: logs, sub: sub}, nil
}

// WatchMasterSnapshottersUpdated is a free log subscription operation binding the contract event 0x9c2d4c2b4cf1ca90e31b1448712dae908d3568785b68a411f6617479c2b9913b.
//
// Solidity: event masterSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) WatchMasterSnapshottersUpdated(opts *bind.WatchOpts, sink chan<- *ContractApiMasterSnapshottersUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractApi.contract.WatchLogs(opts, "masterSnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApiMasterSnapshottersUpdated)
				if err := _ContractApi.contract.UnpackLog(event, "masterSnapshottersUpdated", log); err != nil {
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

// ParseMasterSnapshottersUpdated is a log parse operation binding the contract event 0x9c2d4c2b4cf1ca90e31b1448712dae908d3568785b68a411f6617479c2b9913b.
//
// Solidity: event masterSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ContractApi *ContractApiFilterer) ParseMasterSnapshottersUpdated(log types.Log) (*ContractApiMasterSnapshottersUpdated, error) {
	event := new(ContractApiMasterSnapshottersUpdated)
	if err := _ContractApi.contract.UnpackLog(event, "masterSnapshottersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

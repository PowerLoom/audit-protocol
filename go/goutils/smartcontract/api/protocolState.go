// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractApi

import (
	"errors"
	"math/big"
	"strings"

	"github.com/ava-labs/coreth/accounts/abi"
	"github.com/ava-labs/coreth/accounts/abi/bind"
	"github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = interfaces.NotFound
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

// ProtocolStateMetaData contains all meta data concerning the ProtocolState contract.
var ProtocolStateMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"slotIds\",\"type\":\"uint256[]\"},{\"internalType\":\"address[]\",\"name\":\"snapshotterAddresses\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"timeSlots\",\"type\":\"uint256[]\"}],\"name\":\"assignSnapshotterToSlotsWithTimeSlot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeSlot\",\"type\":\"uint256\"}],\"name\":\"assignSnapshotterToSlotWithTimeSlot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"epochSize\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"sourceChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"sourceChainBlockTime\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"useBlockNumberAsEpochId\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"slotsPerDay\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"InvalidShortString\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"str\",\"type\":\"string\"}],\"name\":\"StringTooLong\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"dayId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DailyTaskCompletedEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"dayId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DayStartedEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DelayedSnapshotSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EIP712DomainChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"EpochReleased\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"name\":\"forceCompleteConsensusSnapshot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"forceSkipEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_dayCounter\",\"type\":\"uint256\"}],\"name\":\"loadCurrentDay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeSlot\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardPoints\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentStreak\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentDaySnapshotCount\",\"type\":\"uint256\"}],\"name\":\"loadSlotInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"dayId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"snapshotCount\",\"type\":\"uint256\"}],\"name\":\"loadSlotSubmissions\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectType\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"enableEpochId\",\"type\":\"uint256\"}],\"name\":\"ProjectTypeUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"releaseEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"finalizedSnapshotCount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalReceivedCount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"snapshotterAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"SnapshotSubmitted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"}],\"internalType\":\"structPowerloomProtocolState.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"submitSnapshot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"toggleRewards\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"toggleTimeSlotCheck\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_projectType\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"_status\",\"type\":\"bool\"}],\"name\":\"updateAllowedProjectType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dailySnapshotQuota\",\"type\":\"uint256\"}],\"name\":\"updateDailySnapshotQuota\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_snapshotters\",\"type\":\"address[]\"},{\"internalType\":\"bool[]\",\"name\":\"_status\",\"type\":\"bool[]\"}],\"name\":\"updateMasterSnapshotters\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minSubmissionsForConsensus\",\"type\":\"uint256\"}],\"name\":\"updateMinSnapshottersForConsensus\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newrewardBasePoints\",\"type\":\"uint256\"}],\"name\":\"updaterewardBasePoints\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newsnapshotSubmissionWindow\",\"type\":\"uint256\"}],\"name\":\"updateSnapshotSubmissionWindow\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newstreakBonusPoints\",\"type\":\"uint256\"}],\"name\":\"updatestreakBonusPoints\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_validators\",\"type\":\"address[]\"},{\"internalType\":\"bool[]\",\"name\":\"_status\",\"type\":\"bool[]\"}],\"name\":\"updateValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"validatorAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"ValidatorsUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"allSnapshottersUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"forceStartDay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"masterSnapshottersUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"allowedProjectTypes\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"allSnapshotters\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"name\":\"checkDynamicConsensusSnapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"day\",\"type\":\"uint256\"}],\"name\":\"checkSlotTaskStatusForDay\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"begin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DailySnapshotQuota\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dayCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DeploymentBlockNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"EPOCH_SIZE\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"epochInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blocknumber\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochEnd\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"epochsInADay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMasterSnapshotters\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"}],\"name\":\"getSlotInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"snapshotterAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeSlot\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardPoints\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentStreak\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentStreakBonus\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentDaySnapshotCount\",\"type\":\"uint256\"}],\"internalType\":\"structPowerloomProtocolState.SlotInfo\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"}],\"name\":\"getSlotStreak\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_slotId\",\"type\":\"uint256\"}],\"name\":\"getSnapshotterTimeSlot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalMasterSnapshotterCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalSnapshotterCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getValidators\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"lastFinalizedSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"lastSnapshotterAddressUpdate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"masterSnapshotters\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxSnapshotsCid\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"maxSnapshotsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minSubmissionsForConsensus\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"projectFirstEpochId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"messageHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"recoverAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardBasePoints\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardsEnabled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"slotCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slotRewardPoints\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SLOTS_PER_DAY\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slotSnapshotterMapping\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slotStreakCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slotSubmissionCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"snapshotsReceived\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"snapshotsReceivedCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"snapshotsReceivedSlot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"snapshotStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"finalized\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"snapshotSubmissionWindow\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SOURCE_CHAIN_BLOCK_TIME\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SOURCE_CHAIN_ID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"streakBonusPoints\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"timeSlotCheck\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"timeSlotPreference\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"totalSnapshotsReceived\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"USE_BLOCK_NUMBER_AS_EPOCH_ID\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"slotId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"snapshotCid\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"epochId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"projectId\",\"type\":\"string\"}],\"internalType\":\"structPowerloomProtocolState.Request\",\"name\":\"request\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"verify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ProtocolStateABI is the input ABI used to generate the binding from.
// Deprecated: Use ProtocolStateMetaData.ABI instead.
var ProtocolStateABI = ProtocolStateMetaData.ABI

// ProtocolState is an auto generated Go binding around an Ethereum contract.
type ProtocolState struct {
	ProtocolStateCaller     // Read-only binding to the contract
	ProtocolStateTransactor // Write-only binding to the contract
	ProtocolStateFilterer   // Log filterer for contract events
}

// ProtocolStateCaller is an auto generated read-only Go binding around an Ethereum contract.
type ProtocolStateCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProtocolStateTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ProtocolStateTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProtocolStateFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ProtocolStateFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProtocolStateSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ProtocolStateSession struct {
	Contract     *ProtocolState    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ProtocolStateCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ProtocolStateCallerSession struct {
	Contract *ProtocolStateCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ProtocolStateTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ProtocolStateTransactorSession struct {
	Contract     *ProtocolStateTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ProtocolStateRaw is an auto generated low-level Go binding around an Ethereum contract.
type ProtocolStateRaw struct {
	Contract *ProtocolState // Generic contract binding to access the raw methods on
}

// ProtocolStateCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ProtocolStateCallerRaw struct {
	Contract *ProtocolStateCaller // Generic read-only contract binding to access the raw methods on
}

// ProtocolStateTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ProtocolStateTransactorRaw struct {
	Contract *ProtocolStateTransactor // Generic write-only contract binding to access the raw methods on
}

// NewProtocolState creates a new instance of ProtocolState, bound to a specific deployed contract.
func NewProtocolState(address common.Address, backend bind.ContractBackend) (*ProtocolState, error) {
	contract, err := bindProtocolState(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ProtocolState{ProtocolStateCaller: ProtocolStateCaller{contract: contract}, ProtocolStateTransactor: ProtocolStateTransactor{contract: contract}, ProtocolStateFilterer: ProtocolStateFilterer{contract: contract}}, nil
}

// NewProtocolStateCaller creates a new read-only instance of ProtocolState, bound to a specific deployed contract.
func NewProtocolStateCaller(address common.Address, caller bind.ContractCaller) (*ProtocolStateCaller, error) {
	contract, err := bindProtocolState(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ProtocolStateCaller{contract: contract}, nil
}

// NewProtocolStateTransactor creates a new write-only instance of ProtocolState, bound to a specific deployed contract.
func NewProtocolStateTransactor(address common.Address, transactor bind.ContractTransactor) (*ProtocolStateTransactor, error) {
	contract, err := bindProtocolState(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ProtocolStateTransactor{contract: contract}, nil
}

// NewProtocolStateFilterer creates a new log filterer instance of ProtocolState, bound to a specific deployed contract.
func NewProtocolStateFilterer(address common.Address, filterer bind.ContractFilterer) (*ProtocolStateFilterer, error) {
	contract, err := bindProtocolState(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ProtocolStateFilterer{contract: contract}, nil
}

// bindProtocolState binds a generic wrapper to an already deployed contract.
func bindProtocolState(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ProtocolStateMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProtocolState *ProtocolStateRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ProtocolState.Contract.ProtocolStateCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProtocolState *ProtocolStateRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProtocolState.Contract.ProtocolStateTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProtocolState *ProtocolStateRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProtocolState.Contract.ProtocolStateTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProtocolState *ProtocolStateCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ProtocolState.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProtocolState *ProtocolStateTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProtocolState.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProtocolState *ProtocolStateTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProtocolState.Contract.contract.Transact(opts, method, params...)
}

// DailySnapshotQuota is a free data retrieval call binding the contract method 0xc7c7077e.
//
// Solidity: function DailySnapshotQuota() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) DailySnapshotQuota(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "DailySnapshotQuota")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DailySnapshotQuota is a free data retrieval call binding the contract method 0xc7c7077e.
//
// Solidity: function DailySnapshotQuota() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) DailySnapshotQuota() (*big.Int, error) {
	return _ProtocolState.Contract.DailySnapshotQuota(&_ProtocolState.CallOpts)
}

// DailySnapshotQuota is a free data retrieval call binding the contract method 0xc7c7077e.
//
// Solidity: function DailySnapshotQuota() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) DailySnapshotQuota() (*big.Int, error) {
	return _ProtocolState.Contract.DailySnapshotQuota(&_ProtocolState.CallOpts)
}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xb1288f71.
//
// Solidity: function DeploymentBlockNumber() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) DeploymentBlockNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "DeploymentBlockNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xb1288f71.
//
// Solidity: function DeploymentBlockNumber() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) DeploymentBlockNumber() (*big.Int, error) {
	return _ProtocolState.Contract.DeploymentBlockNumber(&_ProtocolState.CallOpts)
}

// DeploymentBlockNumber is a free data retrieval call binding the contract method 0xb1288f71.
//
// Solidity: function DeploymentBlockNumber() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) DeploymentBlockNumber() (*big.Int, error) {
	return _ProtocolState.Contract.DeploymentBlockNumber(&_ProtocolState.CallOpts)
}

// EPOCHSIZE is a free data retrieval call binding the contract method 0x62656003.
//
// Solidity: function EPOCH_SIZE() view returns(uint8)
func (_ProtocolState *ProtocolStateCaller) EPOCHSIZE(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "EPOCH_SIZE")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// EPOCHSIZE is a free data retrieval call binding the contract method 0x62656003.
//
// Solidity: function EPOCH_SIZE() view returns(uint8)
func (_ProtocolState *ProtocolStateSession) EPOCHSIZE() (uint8, error) {
	return _ProtocolState.Contract.EPOCHSIZE(&_ProtocolState.CallOpts)
}

// EPOCHSIZE is a free data retrieval call binding the contract method 0x62656003.
//
// Solidity: function EPOCH_SIZE() view returns(uint8)
func (_ProtocolState *ProtocolStateCallerSession) EPOCHSIZE() (uint8, error) {
	return _ProtocolState.Contract.EPOCHSIZE(&_ProtocolState.CallOpts)
}

// SLOTSPERDAY is a free data retrieval call binding the contract method 0xe479b88d.
//
// Solidity: function SLOTS_PER_DAY() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SLOTSPERDAY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "SLOTS_PER_DAY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SLOTSPERDAY is a free data retrieval call binding the contract method 0xe479b88d.
//
// Solidity: function SLOTS_PER_DAY() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SLOTSPERDAY() (*big.Int, error) {
	return _ProtocolState.Contract.SLOTSPERDAY(&_ProtocolState.CallOpts)
}

// SLOTSPERDAY is a free data retrieval call binding the contract method 0xe479b88d.
//
// Solidity: function SLOTS_PER_DAY() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SLOTSPERDAY() (*big.Int, error) {
	return _ProtocolState.Contract.SLOTSPERDAY(&_ProtocolState.CallOpts)
}

// SOURCECHAINBLOCKTIME is a free data retrieval call binding the contract method 0x351b6155.
//
// Solidity: function SOURCE_CHAIN_BLOCK_TIME() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SOURCECHAINBLOCKTIME(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "SOURCE_CHAIN_BLOCK_TIME")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SOURCECHAINBLOCKTIME is a free data retrieval call binding the contract method 0x351b6155.
//
// Solidity: function SOURCE_CHAIN_BLOCK_TIME() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SOURCECHAINBLOCKTIME() (*big.Int, error) {
	return _ProtocolState.Contract.SOURCECHAINBLOCKTIME(&_ProtocolState.CallOpts)
}

// SOURCECHAINBLOCKTIME is a free data retrieval call binding the contract method 0x351b6155.
//
// Solidity: function SOURCE_CHAIN_BLOCK_TIME() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SOURCECHAINBLOCKTIME() (*big.Int, error) {
	return _ProtocolState.Contract.SOURCECHAINBLOCKTIME(&_ProtocolState.CallOpts)
}

// SOURCECHAINID is a free data retrieval call binding the contract method 0x74be2150.
//
// Solidity: function SOURCE_CHAIN_ID() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SOURCECHAINID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "SOURCE_CHAIN_ID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SOURCECHAINID is a free data retrieval call binding the contract method 0x74be2150.
//
// Solidity: function SOURCE_CHAIN_ID() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SOURCECHAINID() (*big.Int, error) {
	return _ProtocolState.Contract.SOURCECHAINID(&_ProtocolState.CallOpts)
}

// SOURCECHAINID is a free data retrieval call binding the contract method 0x74be2150.
//
// Solidity: function SOURCE_CHAIN_ID() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SOURCECHAINID() (*big.Int, error) {
	return _ProtocolState.Contract.SOURCECHAINID(&_ProtocolState.CallOpts)
}

// USEBLOCKNUMBERASEPOCHID is a free data retrieval call binding the contract method 0x2d46247b.
//
// Solidity: function USE_BLOCK_NUMBER_AS_EPOCH_ID() view returns(bool)
func (_ProtocolState *ProtocolStateCaller) USEBLOCKNUMBERASEPOCHID(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "USE_BLOCK_NUMBER_AS_EPOCH_ID")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// USEBLOCKNUMBERASEPOCHID is a free data retrieval call binding the contract method 0x2d46247b.
//
// Solidity: function USE_BLOCK_NUMBER_AS_EPOCH_ID() view returns(bool)
func (_ProtocolState *ProtocolStateSession) USEBLOCKNUMBERASEPOCHID() (bool, error) {
	return _ProtocolState.Contract.USEBLOCKNUMBERASEPOCHID(&_ProtocolState.CallOpts)
}

// USEBLOCKNUMBERASEPOCHID is a free data retrieval call binding the contract method 0x2d46247b.
//
// Solidity: function USE_BLOCK_NUMBER_AS_EPOCH_ID() view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) USEBLOCKNUMBERASEPOCHID() (bool, error) {
	return _ProtocolState.Contract.USEBLOCKNUMBERASEPOCHID(&_ProtocolState.CallOpts)
}

// AllSnapshotters is a free data retrieval call binding the contract method 0x3d15d0f4.
//
// Solidity: function allSnapshotters(address ) view returns(bool)
func (_ProtocolState *ProtocolStateCaller) AllSnapshotters(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "allSnapshotters", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllSnapshotters is a free data retrieval call binding the contract method 0x3d15d0f4.
//
// Solidity: function allSnapshotters(address ) view returns(bool)
func (_ProtocolState *ProtocolStateSession) AllSnapshotters(arg0 common.Address) (bool, error) {
	return _ProtocolState.Contract.AllSnapshotters(&_ProtocolState.CallOpts, arg0)
}

// AllSnapshotters is a free data retrieval call binding the contract method 0x3d15d0f4.
//
// Solidity: function allSnapshotters(address ) view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) AllSnapshotters(arg0 common.Address) (bool, error) {
	return _ProtocolState.Contract.AllSnapshotters(&_ProtocolState.CallOpts, arg0)
}

// AllowedProjectTypes is a free data retrieval call binding the contract method 0x04d76d59.
//
// Solidity: function allowedProjectTypes(string ) view returns(bool)
func (_ProtocolState *ProtocolStateCaller) AllowedProjectTypes(opts *bind.CallOpts, arg0 string) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "allowedProjectTypes", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllowedProjectTypes is a free data retrieval call binding the contract method 0x04d76d59.
//
// Solidity: function allowedProjectTypes(string ) view returns(bool)
func (_ProtocolState *ProtocolStateSession) AllowedProjectTypes(arg0 string) (bool, error) {
	return _ProtocolState.Contract.AllowedProjectTypes(&_ProtocolState.CallOpts, arg0)
}

// AllowedProjectTypes is a free data retrieval call binding the contract method 0x04d76d59.
//
// Solidity: function allowedProjectTypes(string ) view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) AllowedProjectTypes(arg0 string) (bool, error) {
	return _ProtocolState.Contract.AllowedProjectTypes(&_ProtocolState.CallOpts, arg0)
}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochId) view returns(bool)
func (_ProtocolState *ProtocolStateCaller) CheckDynamicConsensusSnapshot(opts *bind.CallOpts, projectId string, epochId *big.Int) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "checkDynamicConsensusSnapshot", projectId, epochId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochId) view returns(bool)
func (_ProtocolState *ProtocolStateSession) CheckDynamicConsensusSnapshot(projectId string, epochId *big.Int) (bool, error) {
	return _ProtocolState.Contract.CheckDynamicConsensusSnapshot(&_ProtocolState.CallOpts, projectId, epochId)
}

// CheckDynamicConsensusSnapshot is a free data retrieval call binding the contract method 0x18b53a54.
//
// Solidity: function checkDynamicConsensusSnapshot(string projectId, uint256 epochId) view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) CheckDynamicConsensusSnapshot(projectId string, epochId *big.Int) (bool, error) {
	return _ProtocolState.Contract.CheckDynamicConsensusSnapshot(&_ProtocolState.CallOpts, projectId, epochId)
}

// CheckSlotTaskStatusForDay is a free data retrieval call binding the contract method 0xd1dd6ddd.
//
// Solidity: function checkSlotTaskStatusForDay(uint256 slotId, uint256 day) view returns(bool)
func (_ProtocolState *ProtocolStateCaller) CheckSlotTaskStatusForDay(opts *bind.CallOpts, slotId *big.Int, day *big.Int) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "checkSlotTaskStatusForDay", slotId, day)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckSlotTaskStatusForDay is a free data retrieval call binding the contract method 0xd1dd6ddd.
//
// Solidity: function checkSlotTaskStatusForDay(uint256 slotId, uint256 day) view returns(bool)
func (_ProtocolState *ProtocolStateSession) CheckSlotTaskStatusForDay(slotId *big.Int, day *big.Int) (bool, error) {
	return _ProtocolState.Contract.CheckSlotTaskStatusForDay(&_ProtocolState.CallOpts, slotId, day)
}

// CheckSlotTaskStatusForDay is a free data retrieval call binding the contract method 0xd1dd6ddd.
//
// Solidity: function checkSlotTaskStatusForDay(uint256 slotId, uint256 day) view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) CheckSlotTaskStatusForDay(slotId *big.Int, day *big.Int) (bool, error) {
	return _ProtocolState.Contract.CheckSlotTaskStatusForDay(&_ProtocolState.CallOpts, slotId, day)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256 begin, uint256 end, uint256 epochId)
func (_ProtocolState *ProtocolStateCaller) CurrentEpoch(opts *bind.CallOpts) (struct {
	Begin   *big.Int
	End     *big.Int
	EpochId *big.Int
}, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "currentEpoch")

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
func (_ProtocolState *ProtocolStateSession) CurrentEpoch() (struct {
	Begin   *big.Int
	End     *big.Int
	EpochId *big.Int
}, error) {
	return _ProtocolState.Contract.CurrentEpoch(&_ProtocolState.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256 begin, uint256 end, uint256 epochId)
func (_ProtocolState *ProtocolStateCallerSession) CurrentEpoch() (struct {
	Begin   *big.Int
	End     *big.Int
	EpochId *big.Int
}, error) {
	return _ProtocolState.Contract.CurrentEpoch(&_ProtocolState.CallOpts)
}

// DayCounter is a free data retrieval call binding the contract method 0x99332c5e.
//
// Solidity: function dayCounter() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) DayCounter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "dayCounter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DayCounter is a free data retrieval call binding the contract method 0x99332c5e.
//
// Solidity: function dayCounter() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) DayCounter() (*big.Int, error) {
	return _ProtocolState.Contract.DayCounter(&_ProtocolState.CallOpts)
}

// DayCounter is a free data retrieval call binding the contract method 0x99332c5e.
//
// Solidity: function dayCounter() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) DayCounter() (*big.Int, error) {
	return _ProtocolState.Contract.DayCounter(&_ProtocolState.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_ProtocolState *ProtocolStateCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "eip712Domain")

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
func (_ProtocolState *ProtocolStateSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _ProtocolState.Contract.Eip712Domain(&_ProtocolState.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_ProtocolState *ProtocolStateCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _ProtocolState.Contract.Eip712Domain(&_ProtocolState.CallOpts)
}

// EpochInfo is a free data retrieval call binding the contract method 0x3894228e.
//
// Solidity: function epochInfo(uint256 ) view returns(uint256 timestamp, uint256 blocknumber, uint256 epochEnd)
func (_ProtocolState *ProtocolStateCaller) EpochInfo(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
	EpochEnd    *big.Int
}, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "epochInfo", arg0)

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
func (_ProtocolState *ProtocolStateSession) EpochInfo(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
	EpochEnd    *big.Int
}, error) {
	return _ProtocolState.Contract.EpochInfo(&_ProtocolState.CallOpts, arg0)
}

// EpochInfo is a free data retrieval call binding the contract method 0x3894228e.
//
// Solidity: function epochInfo(uint256 ) view returns(uint256 timestamp, uint256 blocknumber, uint256 epochEnd)
func (_ProtocolState *ProtocolStateCallerSession) EpochInfo(arg0 *big.Int) (struct {
	Timestamp   *big.Int
	Blocknumber *big.Int
	EpochEnd    *big.Int
}, error) {
	return _ProtocolState.Contract.EpochInfo(&_ProtocolState.CallOpts, arg0)
}

// EpochsInADay is a free data retrieval call binding the contract method 0xe3042b17.
//
// Solidity: function epochsInADay() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) EpochsInADay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "epochsInADay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EpochsInADay is a free data retrieval call binding the contract method 0xe3042b17.
//
// Solidity: function epochsInADay() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) EpochsInADay() (*big.Int, error) {
	return _ProtocolState.Contract.EpochsInADay(&_ProtocolState.CallOpts)
}

// EpochsInADay is a free data retrieval call binding the contract method 0xe3042b17.
//
// Solidity: function epochsInADay() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) EpochsInADay() (*big.Int, error) {
	return _ProtocolState.Contract.EpochsInADay(&_ProtocolState.CallOpts)
}

// GetMasterSnapshotters is a free data retrieval call binding the contract method 0x90110313.
//
// Solidity: function getMasterSnapshotters() view returns(address[])
func (_ProtocolState *ProtocolStateCaller) GetMasterSnapshotters(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "getMasterSnapshotters")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetMasterSnapshotters is a free data retrieval call binding the contract method 0x90110313.
//
// Solidity: function getMasterSnapshotters() view returns(address[])
func (_ProtocolState *ProtocolStateSession) GetMasterSnapshotters() ([]common.Address, error) {
	return _ProtocolState.Contract.GetMasterSnapshotters(&_ProtocolState.CallOpts)
}

// GetMasterSnapshotters is a free data retrieval call binding the contract method 0x90110313.
//
// Solidity: function getMasterSnapshotters() view returns(address[])
func (_ProtocolState *ProtocolStateCallerSession) GetMasterSnapshotters() ([]common.Address, error) {
	return _ProtocolState.Contract.GetMasterSnapshotters(&_ProtocolState.CallOpts)
}

// GetSlotInfo is a free data retrieval call binding the contract method 0xbe20f9ac.
//
// Solidity: function getSlotInfo(uint256 slotId) view returns((uint256,address,uint256,uint256,uint256,uint256,uint256))
func (_ProtocolState *ProtocolStateCaller) GetSlotInfo(opts *bind.CallOpts, slotId *big.Int) (PowerloomProtocolStateSlotInfo, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "getSlotInfo", slotId)

	if err != nil {
		return *new(PowerloomProtocolStateSlotInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(PowerloomProtocolStateSlotInfo)).(*PowerloomProtocolStateSlotInfo)

	return out0, err

}

// GetSlotInfo is a free data retrieval call binding the contract method 0xbe20f9ac.
//
// Solidity: function getSlotInfo(uint256 slotId) view returns((uint256,address,uint256,uint256,uint256,uint256,uint256))
func (_ProtocolState *ProtocolStateSession) GetSlotInfo(slotId *big.Int) (PowerloomProtocolStateSlotInfo, error) {
	return _ProtocolState.Contract.GetSlotInfo(&_ProtocolState.CallOpts, slotId)
}

// GetSlotInfo is a free data retrieval call binding the contract method 0xbe20f9ac.
//
// Solidity: function getSlotInfo(uint256 slotId) view returns((uint256,address,uint256,uint256,uint256,uint256,uint256))
func (_ProtocolState *ProtocolStateCallerSession) GetSlotInfo(slotId *big.Int) (PowerloomProtocolStateSlotInfo, error) {
	return _ProtocolState.Contract.GetSlotInfo(&_ProtocolState.CallOpts, slotId)
}

// GetSlotStreak is a free data retrieval call binding the contract method 0x2252c7b8.
//
// Solidity: function getSlotStreak(uint256 slotId) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) GetSlotStreak(opts *bind.CallOpts, slotId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "getSlotStreak", slotId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlotStreak is a free data retrieval call binding the contract method 0x2252c7b8.
//
// Solidity: function getSlotStreak(uint256 slotId) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) GetSlotStreak(slotId *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.GetSlotStreak(&_ProtocolState.CallOpts, slotId)
}

// GetSlotStreak is a free data retrieval call binding the contract method 0x2252c7b8.
//
// Solidity: function getSlotStreak(uint256 slotId) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) GetSlotStreak(slotId *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.GetSlotStreak(&_ProtocolState.CallOpts, slotId)
}

// GetSnapshotterTimeSlot is a free data retrieval call binding the contract method 0x2db9ace2.
//
// Solidity: function getSnapshotterTimeSlot(uint256 _slotId) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) GetSnapshotterTimeSlot(opts *bind.CallOpts, _slotId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "getSnapshotterTimeSlot", _slotId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSnapshotterTimeSlot is a free data retrieval call binding the contract method 0x2db9ace2.
//
// Solidity: function getSnapshotterTimeSlot(uint256 _slotId) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) GetSnapshotterTimeSlot(_slotId *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.GetSnapshotterTimeSlot(&_ProtocolState.CallOpts, _slotId)
}

// GetSnapshotterTimeSlot is a free data retrieval call binding the contract method 0x2db9ace2.
//
// Solidity: function getSnapshotterTimeSlot(uint256 _slotId) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) GetSnapshotterTimeSlot(_slotId *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.GetSnapshotterTimeSlot(&_ProtocolState.CallOpts, _slotId)
}

// GetTotalMasterSnapshotterCount is a free data retrieval call binding the contract method 0xd551672b.
//
// Solidity: function getTotalMasterSnapshotterCount() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) GetTotalMasterSnapshotterCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "getTotalMasterSnapshotterCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalMasterSnapshotterCount is a free data retrieval call binding the contract method 0xd551672b.
//
// Solidity: function getTotalMasterSnapshotterCount() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) GetTotalMasterSnapshotterCount() (*big.Int, error) {
	return _ProtocolState.Contract.GetTotalMasterSnapshotterCount(&_ProtocolState.CallOpts)
}

// GetTotalMasterSnapshotterCount is a free data retrieval call binding the contract method 0xd551672b.
//
// Solidity: function getTotalMasterSnapshotterCount() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) GetTotalMasterSnapshotterCount() (*big.Int, error) {
	return _ProtocolState.Contract.GetTotalMasterSnapshotterCount(&_ProtocolState.CallOpts)
}

// GetTotalSnapshotterCount is a free data retrieval call binding the contract method 0x92ae6f66.
//
// Solidity: function getTotalSnapshotterCount() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) GetTotalSnapshotterCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "getTotalSnapshotterCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalSnapshotterCount is a free data retrieval call binding the contract method 0x92ae6f66.
//
// Solidity: function getTotalSnapshotterCount() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) GetTotalSnapshotterCount() (*big.Int, error) {
	return _ProtocolState.Contract.GetTotalSnapshotterCount(&_ProtocolState.CallOpts)
}

// GetTotalSnapshotterCount is a free data retrieval call binding the contract method 0x92ae6f66.
//
// Solidity: function getTotalSnapshotterCount() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) GetTotalSnapshotterCount() (*big.Int, error) {
	return _ProtocolState.Contract.GetTotalSnapshotterCount(&_ProtocolState.CallOpts)
}

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_ProtocolState *ProtocolStateCaller) GetValidators(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "getValidators")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_ProtocolState *ProtocolStateSession) GetValidators() ([]common.Address, error) {
	return _ProtocolState.Contract.GetValidators(&_ProtocolState.CallOpts)
}

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_ProtocolState *ProtocolStateCallerSession) GetValidators() ([]common.Address, error) {
	return _ProtocolState.Contract.GetValidators(&_ProtocolState.CallOpts)
}

// LastFinalizedSnapshot is a free data retrieval call binding the contract method 0x4ea16b0a.
//
// Solidity: function lastFinalizedSnapshot(string ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) LastFinalizedSnapshot(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "lastFinalizedSnapshot", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastFinalizedSnapshot is a free data retrieval call binding the contract method 0x4ea16b0a.
//
// Solidity: function lastFinalizedSnapshot(string ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) LastFinalizedSnapshot(arg0 string) (*big.Int, error) {
	return _ProtocolState.Contract.LastFinalizedSnapshot(&_ProtocolState.CallOpts, arg0)
}

// LastFinalizedSnapshot is a free data retrieval call binding the contract method 0x4ea16b0a.
//
// Solidity: function lastFinalizedSnapshot(string ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) LastFinalizedSnapshot(arg0 string) (*big.Int, error) {
	return _ProtocolState.Contract.LastFinalizedSnapshot(&_ProtocolState.CallOpts, arg0)
}

// LastSnapshotterAddressUpdate is a free data retrieval call binding the contract method 0x25d6ad01.
//
// Solidity: function lastSnapshotterAddressUpdate(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) LastSnapshotterAddressUpdate(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "lastSnapshotterAddressUpdate", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastSnapshotterAddressUpdate is a free data retrieval call binding the contract method 0x25d6ad01.
//
// Solidity: function lastSnapshotterAddressUpdate(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) LastSnapshotterAddressUpdate(arg0 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.LastSnapshotterAddressUpdate(&_ProtocolState.CallOpts, arg0)
}

// LastSnapshotterAddressUpdate is a free data retrieval call binding the contract method 0x25d6ad01.
//
// Solidity: function lastSnapshotterAddressUpdate(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) LastSnapshotterAddressUpdate(arg0 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.LastSnapshotterAddressUpdate(&_ProtocolState.CallOpts, arg0)
}

// MasterSnapshotters is a free data retrieval call binding the contract method 0x34b739d9.
//
// Solidity: function masterSnapshotters(address ) view returns(bool)
func (_ProtocolState *ProtocolStateCaller) MasterSnapshotters(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "masterSnapshotters", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MasterSnapshotters is a free data retrieval call binding the contract method 0x34b739d9.
//
// Solidity: function masterSnapshotters(address ) view returns(bool)
func (_ProtocolState *ProtocolStateSession) MasterSnapshotters(arg0 common.Address) (bool, error) {
	return _ProtocolState.Contract.MasterSnapshotters(&_ProtocolState.CallOpts, arg0)
}

// MasterSnapshotters is a free data retrieval call binding the contract method 0x34b739d9.
//
// Solidity: function masterSnapshotters(address ) view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) MasterSnapshotters(arg0 common.Address) (bool, error) {
	return _ProtocolState.Contract.MasterSnapshotters(&_ProtocolState.CallOpts, arg0)
}

// MaxSnapshotsCid is a free data retrieval call binding the contract method 0xc2b97d4c.
//
// Solidity: function maxSnapshotsCid(string , uint256 ) view returns(string)
func (_ProtocolState *ProtocolStateCaller) MaxSnapshotsCid(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (string, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "maxSnapshotsCid", arg0, arg1)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// MaxSnapshotsCid is a free data retrieval call binding the contract method 0xc2b97d4c.
//
// Solidity: function maxSnapshotsCid(string , uint256 ) view returns(string)
func (_ProtocolState *ProtocolStateSession) MaxSnapshotsCid(arg0 string, arg1 *big.Int) (string, error) {
	return _ProtocolState.Contract.MaxSnapshotsCid(&_ProtocolState.CallOpts, arg0, arg1)
}

// MaxSnapshotsCid is a free data retrieval call binding the contract method 0xc2b97d4c.
//
// Solidity: function maxSnapshotsCid(string , uint256 ) view returns(string)
func (_ProtocolState *ProtocolStateCallerSession) MaxSnapshotsCid(arg0 string, arg1 *big.Int) (string, error) {
	return _ProtocolState.Contract.MaxSnapshotsCid(&_ProtocolState.CallOpts, arg0, arg1)
}

// MaxSnapshotsCount is a free data retrieval call binding the contract method 0xdbb54182.
//
// Solidity: function maxSnapshotsCount(string , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) MaxSnapshotsCount(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "maxSnapshotsCount", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxSnapshotsCount is a free data retrieval call binding the contract method 0xdbb54182.
//
// Solidity: function maxSnapshotsCount(string , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) MaxSnapshotsCount(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.MaxSnapshotsCount(&_ProtocolState.CallOpts, arg0, arg1)
}

// MaxSnapshotsCount is a free data retrieval call binding the contract method 0xdbb54182.
//
// Solidity: function maxSnapshotsCount(string , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) MaxSnapshotsCount(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.MaxSnapshotsCount(&_ProtocolState.CallOpts, arg0, arg1)
}

// MinSubmissionsForConsensus is a free data retrieval call binding the contract method 0x66752e1e.
//
// Solidity: function minSubmissionsForConsensus() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) MinSubmissionsForConsensus(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "minSubmissionsForConsensus")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinSubmissionsForConsensus is a free data retrieval call binding the contract method 0x66752e1e.
//
// Solidity: function minSubmissionsForConsensus() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) MinSubmissionsForConsensus() (*big.Int, error) {
	return _ProtocolState.Contract.MinSubmissionsForConsensus(&_ProtocolState.CallOpts)
}

// MinSubmissionsForConsensus is a free data retrieval call binding the contract method 0x66752e1e.
//
// Solidity: function minSubmissionsForConsensus() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) MinSubmissionsForConsensus() (*big.Int, error) {
	return _ProtocolState.Contract.MinSubmissionsForConsensus(&_ProtocolState.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProtocolState *ProtocolStateCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProtocolState *ProtocolStateSession) Owner() (common.Address, error) {
	return _ProtocolState.Contract.Owner(&_ProtocolState.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProtocolState *ProtocolStateCallerSession) Owner() (common.Address, error) {
	return _ProtocolState.Contract.Owner(&_ProtocolState.CallOpts)
}

// ProjectFirstEpochId is a free data retrieval call binding the contract method 0xfa30dbe0.
//
// Solidity: function projectFirstEpochId(string ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) ProjectFirstEpochId(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "projectFirstEpochId", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProjectFirstEpochId is a free data retrieval call binding the contract method 0xfa30dbe0.
//
// Solidity: function projectFirstEpochId(string ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) ProjectFirstEpochId(arg0 string) (*big.Int, error) {
	return _ProtocolState.Contract.ProjectFirstEpochId(&_ProtocolState.CallOpts, arg0)
}

// ProjectFirstEpochId is a free data retrieval call binding the contract method 0xfa30dbe0.
//
// Solidity: function projectFirstEpochId(string ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) ProjectFirstEpochId(arg0 string) (*big.Int, error) {
	return _ProtocolState.Contract.ProjectFirstEpochId(&_ProtocolState.CallOpts, arg0)
}

// RecoverAddress is a free data retrieval call binding the contract method 0xc655d7aa.
//
// Solidity: function recoverAddress(bytes32 messageHash, bytes signature) pure returns(address)
func (_ProtocolState *ProtocolStateCaller) RecoverAddress(opts *bind.CallOpts, messageHash [32]byte, signature []byte) (common.Address, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "recoverAddress", messageHash, signature)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RecoverAddress is a free data retrieval call binding the contract method 0xc655d7aa.
//
// Solidity: function recoverAddress(bytes32 messageHash, bytes signature) pure returns(address)
func (_ProtocolState *ProtocolStateSession) RecoverAddress(messageHash [32]byte, signature []byte) (common.Address, error) {
	return _ProtocolState.Contract.RecoverAddress(&_ProtocolState.CallOpts, messageHash, signature)
}

// RecoverAddress is a free data retrieval call binding the contract method 0xc655d7aa.
//
// Solidity: function recoverAddress(bytes32 messageHash, bytes signature) pure returns(address)
func (_ProtocolState *ProtocolStateCallerSession) RecoverAddress(messageHash [32]byte, signature []byte) (common.Address, error) {
	return _ProtocolState.Contract.RecoverAddress(&_ProtocolState.CallOpts, messageHash, signature)
}

// RewardBasePoints is a free data retrieval call binding the contract method 0x7e9834d7.
//
// Solidity: function rewardBasePoints() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) RewardBasePoints(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "rewardBasePoints")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardBasePoints is a free data retrieval call binding the contract method 0x7e9834d7.
//
// Solidity: function rewardBasePoints() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) RewardBasePoints() (*big.Int, error) {
	return _ProtocolState.Contract.RewardBasePoints(&_ProtocolState.CallOpts)
}

// RewardBasePoints is a free data retrieval call binding the contract method 0x7e9834d7.
//
// Solidity: function rewardBasePoints() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) RewardBasePoints() (*big.Int, error) {
	return _ProtocolState.Contract.RewardBasePoints(&_ProtocolState.CallOpts)
}

// RewardsEnabled is a free data retrieval call binding the contract method 0x1dafe16b.
//
// Solidity: function rewardsEnabled() view returns(bool)
func (_ProtocolState *ProtocolStateCaller) RewardsEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "rewardsEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RewardsEnabled is a free data retrieval call binding the contract method 0x1dafe16b.
//
// Solidity: function rewardsEnabled() view returns(bool)
func (_ProtocolState *ProtocolStateSession) RewardsEnabled() (bool, error) {
	return _ProtocolState.Contract.RewardsEnabled(&_ProtocolState.CallOpts)
}

// RewardsEnabled is a free data retrieval call binding the contract method 0x1dafe16b.
//
// Solidity: function rewardsEnabled() view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) RewardsEnabled() (bool, error) {
	return _ProtocolState.Contract.RewardsEnabled(&_ProtocolState.CallOpts)
}

// SlotCounter is a free data retrieval call binding the contract method 0xe59a4105.
//
// Solidity: function slotCounter() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SlotCounter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "slotCounter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotCounter is a free data retrieval call binding the contract method 0xe59a4105.
//
// Solidity: function slotCounter() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SlotCounter() (*big.Int, error) {
	return _ProtocolState.Contract.SlotCounter(&_ProtocolState.CallOpts)
}

// SlotCounter is a free data retrieval call binding the contract method 0xe59a4105.
//
// Solidity: function slotCounter() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SlotCounter() (*big.Int, error) {
	return _ProtocolState.Contract.SlotCounter(&_ProtocolState.CallOpts)
}

// SlotRewardPoints is a free data retrieval call binding the contract method 0x486429e7.
//
// Solidity: function slotRewardPoints(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SlotRewardPoints(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "slotRewardPoints", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotRewardPoints is a free data retrieval call binding the contract method 0x486429e7.
//
// Solidity: function slotRewardPoints(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SlotRewardPoints(arg0 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.SlotRewardPoints(&_ProtocolState.CallOpts, arg0)
}

// SlotRewardPoints is a free data retrieval call binding the contract method 0x486429e7.
//
// Solidity: function slotRewardPoints(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SlotRewardPoints(arg0 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.SlotRewardPoints(&_ProtocolState.CallOpts, arg0)
}

// SlotSnapshotterMapping is a free data retrieval call binding the contract method 0x948a463e.
//
// Solidity: function slotSnapshotterMapping(uint256 ) view returns(address)
func (_ProtocolState *ProtocolStateCaller) SlotSnapshotterMapping(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "slotSnapshotterMapping", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SlotSnapshotterMapping is a free data retrieval call binding the contract method 0x948a463e.
//
// Solidity: function slotSnapshotterMapping(uint256 ) view returns(address)
func (_ProtocolState *ProtocolStateSession) SlotSnapshotterMapping(arg0 *big.Int) (common.Address, error) {
	return _ProtocolState.Contract.SlotSnapshotterMapping(&_ProtocolState.CallOpts, arg0)
}

// SlotSnapshotterMapping is a free data retrieval call binding the contract method 0x948a463e.
//
// Solidity: function slotSnapshotterMapping(uint256 ) view returns(address)
func (_ProtocolState *ProtocolStateCallerSession) SlotSnapshotterMapping(arg0 *big.Int) (common.Address, error) {
	return _ProtocolState.Contract.SlotSnapshotterMapping(&_ProtocolState.CallOpts, arg0)
}

// SlotStreakCounter is a free data retrieval call binding the contract method 0x7c5f1557.
//
// Solidity: function slotStreakCounter(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SlotStreakCounter(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "slotStreakCounter", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotStreakCounter is a free data retrieval call binding the contract method 0x7c5f1557.
//
// Solidity: function slotStreakCounter(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SlotStreakCounter(arg0 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.SlotStreakCounter(&_ProtocolState.CallOpts, arg0)
}

// SlotStreakCounter is a free data retrieval call binding the contract method 0x7c5f1557.
//
// Solidity: function slotStreakCounter(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SlotStreakCounter(arg0 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.SlotStreakCounter(&_ProtocolState.CallOpts, arg0)
}

// SlotSubmissionCount is a free data retrieval call binding the contract method 0x9dbc5064.
//
// Solidity: function slotSubmissionCount(uint256 , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SlotSubmissionCount(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "slotSubmissionCount", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlotSubmissionCount is a free data retrieval call binding the contract method 0x9dbc5064.
//
// Solidity: function slotSubmissionCount(uint256 , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SlotSubmissionCount(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.SlotSubmissionCount(&_ProtocolState.CallOpts, arg0, arg1)
}

// SlotSubmissionCount is a free data retrieval call binding the contract method 0x9dbc5064.
//
// Solidity: function slotSubmissionCount(uint256 , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SlotSubmissionCount(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.SlotSubmissionCount(&_ProtocolState.CallOpts, arg0, arg1)
}

// SnapshotStatus is a free data retrieval call binding the contract method 0x3aaf384d.
//
// Solidity: function snapshotStatus(string , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ProtocolState *ProtocolStateCaller) SnapshotStatus(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "snapshotStatus", arg0, arg1)

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
func (_ProtocolState *ProtocolStateSession) SnapshotStatus(arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	return _ProtocolState.Contract.SnapshotStatus(&_ProtocolState.CallOpts, arg0, arg1)
}

// SnapshotStatus is a free data retrieval call binding the contract method 0x3aaf384d.
//
// Solidity: function snapshotStatus(string , uint256 ) view returns(bool finalized, uint256 timestamp)
func (_ProtocolState *ProtocolStateCallerSession) SnapshotStatus(arg0 string, arg1 *big.Int) (struct {
	Finalized bool
	Timestamp *big.Int
}, error) {
	return _ProtocolState.Contract.SnapshotStatus(&_ProtocolState.CallOpts, arg0, arg1)
}

// SnapshotSubmissionWindow is a free data retrieval call binding the contract method 0x059080f6.
//
// Solidity: function snapshotSubmissionWindow() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SnapshotSubmissionWindow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "snapshotSubmissionWindow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SnapshotSubmissionWindow is a free data retrieval call binding the contract method 0x059080f6.
//
// Solidity: function snapshotSubmissionWindow() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SnapshotSubmissionWindow() (*big.Int, error) {
	return _ProtocolState.Contract.SnapshotSubmissionWindow(&_ProtocolState.CallOpts)
}

// SnapshotSubmissionWindow is a free data retrieval call binding the contract method 0x059080f6.
//
// Solidity: function snapshotSubmissionWindow() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SnapshotSubmissionWindow() (*big.Int, error) {
	return _ProtocolState.Contract.SnapshotSubmissionWindow(&_ProtocolState.CallOpts)
}

// SnapshotsReceived is a free data retrieval call binding the contract method 0x04dbb717.
//
// Solidity: function snapshotsReceived(string , uint256 , address ) view returns(bool)
func (_ProtocolState *ProtocolStateCaller) SnapshotsReceived(opts *bind.CallOpts, arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "snapshotsReceived", arg0, arg1, arg2)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SnapshotsReceived is a free data retrieval call binding the contract method 0x04dbb717.
//
// Solidity: function snapshotsReceived(string , uint256 , address ) view returns(bool)
func (_ProtocolState *ProtocolStateSession) SnapshotsReceived(arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	return _ProtocolState.Contract.SnapshotsReceived(&_ProtocolState.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceived is a free data retrieval call binding the contract method 0x04dbb717.
//
// Solidity: function snapshotsReceived(string , uint256 , address ) view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) SnapshotsReceived(arg0 string, arg1 *big.Int, arg2 common.Address) (bool, error) {
	return _ProtocolState.Contract.SnapshotsReceived(&_ProtocolState.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceivedCount is a free data retrieval call binding the contract method 0x8e07ba42.
//
// Solidity: function snapshotsReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) SnapshotsReceivedCount(opts *bind.CallOpts, arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "snapshotsReceivedCount", arg0, arg1, arg2)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SnapshotsReceivedCount is a free data retrieval call binding the contract method 0x8e07ba42.
//
// Solidity: function snapshotsReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) SnapshotsReceivedCount(arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	return _ProtocolState.Contract.SnapshotsReceivedCount(&_ProtocolState.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceivedCount is a free data retrieval call binding the contract method 0x8e07ba42.
//
// Solidity: function snapshotsReceivedCount(string , uint256 , string ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) SnapshotsReceivedCount(arg0 string, arg1 *big.Int, arg2 string) (*big.Int, error) {
	return _ProtocolState.Contract.SnapshotsReceivedCount(&_ProtocolState.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceivedSlot is a free data retrieval call binding the contract method 0x37d080f0.
//
// Solidity: function snapshotsReceivedSlot(string , uint256 , uint256 ) view returns(bool)
func (_ProtocolState *ProtocolStateCaller) SnapshotsReceivedSlot(opts *bind.CallOpts, arg0 string, arg1 *big.Int, arg2 *big.Int) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "snapshotsReceivedSlot", arg0, arg1, arg2)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SnapshotsReceivedSlot is a free data retrieval call binding the contract method 0x37d080f0.
//
// Solidity: function snapshotsReceivedSlot(string , uint256 , uint256 ) view returns(bool)
func (_ProtocolState *ProtocolStateSession) SnapshotsReceivedSlot(arg0 string, arg1 *big.Int, arg2 *big.Int) (bool, error) {
	return _ProtocolState.Contract.SnapshotsReceivedSlot(&_ProtocolState.CallOpts, arg0, arg1, arg2)
}

// SnapshotsReceivedSlot is a free data retrieval call binding the contract method 0x37d080f0.
//
// Solidity: function snapshotsReceivedSlot(string , uint256 , uint256 ) view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) SnapshotsReceivedSlot(arg0 string, arg1 *big.Int, arg2 *big.Int) (bool, error) {
	return _ProtocolState.Contract.SnapshotsReceivedSlot(&_ProtocolState.CallOpts, arg0, arg1, arg2)
}

// StreakBonusPoints is a free data retrieval call binding the contract method 0x36f71682.
//
// Solidity: function streakBonusPoints() view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) StreakBonusPoints(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "streakBonusPoints")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StreakBonusPoints is a free data retrieval call binding the contract method 0x36f71682.
//
// Solidity: function streakBonusPoints() view returns(uint256)
func (_ProtocolState *ProtocolStateSession) StreakBonusPoints() (*big.Int, error) {
	return _ProtocolState.Contract.StreakBonusPoints(&_ProtocolState.CallOpts)
}

// StreakBonusPoints is a free data retrieval call binding the contract method 0x36f71682.
//
// Solidity: function streakBonusPoints() view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) StreakBonusPoints() (*big.Int, error) {
	return _ProtocolState.Contract.StreakBonusPoints(&_ProtocolState.CallOpts)
}

// TimeSlotCheck is a free data retrieval call binding the contract method 0x1a0adc24.
//
// Solidity: function timeSlotCheck() view returns(bool)
func (_ProtocolState *ProtocolStateCaller) TimeSlotCheck(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "timeSlotCheck")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TimeSlotCheck is a free data retrieval call binding the contract method 0x1a0adc24.
//
// Solidity: function timeSlotCheck() view returns(bool)
func (_ProtocolState *ProtocolStateSession) TimeSlotCheck() (bool, error) {
	return _ProtocolState.Contract.TimeSlotCheck(&_ProtocolState.CallOpts)
}

// TimeSlotCheck is a free data retrieval call binding the contract method 0x1a0adc24.
//
// Solidity: function timeSlotCheck() view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) TimeSlotCheck() (bool, error) {
	return _ProtocolState.Contract.TimeSlotCheck(&_ProtocolState.CallOpts)
}

// TimeSlotPreference is a free data retrieval call binding the contract method 0x1cdd886b.
//
// Solidity: function timeSlotPreference(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) TimeSlotPreference(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "timeSlotPreference", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TimeSlotPreference is a free data retrieval call binding the contract method 0x1cdd886b.
//
// Solidity: function timeSlotPreference(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) TimeSlotPreference(arg0 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.TimeSlotPreference(&_ProtocolState.CallOpts, arg0)
}

// TimeSlotPreference is a free data retrieval call binding the contract method 0x1cdd886b.
//
// Solidity: function timeSlotPreference(uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) TimeSlotPreference(arg0 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.TimeSlotPreference(&_ProtocolState.CallOpts, arg0)
}

// TotalSnapshotsReceived is a free data retrieval call binding the contract method 0x797bbd58.
//
// Solidity: function totalSnapshotsReceived(string , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCaller) TotalSnapshotsReceived(opts *bind.CallOpts, arg0 string, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "totalSnapshotsReceived", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSnapshotsReceived is a free data retrieval call binding the contract method 0x797bbd58.
//
// Solidity: function totalSnapshotsReceived(string , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateSession) TotalSnapshotsReceived(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.TotalSnapshotsReceived(&_ProtocolState.CallOpts, arg0, arg1)
}

// TotalSnapshotsReceived is a free data retrieval call binding the contract method 0x797bbd58.
//
// Solidity: function totalSnapshotsReceived(string , uint256 ) view returns(uint256)
func (_ProtocolState *ProtocolStateCallerSession) TotalSnapshotsReceived(arg0 string, arg1 *big.Int) (*big.Int, error) {
	return _ProtocolState.Contract.TotalSnapshotsReceived(&_ProtocolState.CallOpts, arg0, arg1)
}

// Verify is a free data retrieval call binding the contract method 0x58939f83.
//
// Solidity: function verify(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ProtocolState *ProtocolStateCaller) Verify(opts *bind.CallOpts, slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	var out []interface{}
	err := _ProtocolState.contract.Call(opts, &out, "verify", slotId, snapshotCid, epochId, projectId, request, signature, signer)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Verify is a free data retrieval call binding the contract method 0x58939f83.
//
// Solidity: function verify(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ProtocolState *ProtocolStateSession) Verify(slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	return _ProtocolState.Contract.Verify(&_ProtocolState.CallOpts, slotId, snapshotCid, epochId, projectId, request, signature, signer)
}

// Verify is a free data retrieval call binding the contract method 0x58939f83.
//
// Solidity: function verify(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature, address signer) view returns(bool)
func (_ProtocolState *ProtocolStateCallerSession) Verify(slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte, signer common.Address) (bool, error) {
	return _ProtocolState.Contract.Verify(&_ProtocolState.CallOpts, slotId, snapshotCid, epochId, projectId, request, signature, signer)
}

// AssignSnapshotterToSlotWithTimeSlot is a paid mutator transaction binding the contract method 0x7341fd46.
//
// Solidity: function assignSnapshotterToSlotWithTimeSlot(uint256 slotId, address snapshotterAddress, uint256 timeSlot) returns()
func (_ProtocolState *ProtocolStateTransactor) AssignSnapshotterToSlotWithTimeSlot(opts *bind.TransactOpts, slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "assignSnapshotterToSlotWithTimeSlot", slotId, snapshotterAddress, timeSlot)
}

// AssignSnapshotterToSlotWithTimeSlot is a paid mutator transaction binding the contract method 0x7341fd46.
//
// Solidity: function assignSnapshotterToSlotWithTimeSlot(uint256 slotId, address snapshotterAddress, uint256 timeSlot) returns()
func (_ProtocolState *ProtocolStateSession) AssignSnapshotterToSlotWithTimeSlot(slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.AssignSnapshotterToSlotWithTimeSlot(&_ProtocolState.TransactOpts, slotId, snapshotterAddress, timeSlot)
}

// AssignSnapshotterToSlotWithTimeSlot is a paid mutator transaction binding the contract method 0x7341fd46.
//
// Solidity: function assignSnapshotterToSlotWithTimeSlot(uint256 slotId, address snapshotterAddress, uint256 timeSlot) returns()
func (_ProtocolState *ProtocolStateTransactorSession) AssignSnapshotterToSlotWithTimeSlot(slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.AssignSnapshotterToSlotWithTimeSlot(&_ProtocolState.TransactOpts, slotId, snapshotterAddress, timeSlot)
}

// AssignSnapshotterToSlotsWithTimeSlot is a paid mutator transaction binding the contract method 0xcd6d56bd.
//
// Solidity: function assignSnapshotterToSlotsWithTimeSlot(uint256[] slotIds, address[] snapshotterAddresses, uint256[] timeSlots) returns()
func (_ProtocolState *ProtocolStateTransactor) AssignSnapshotterToSlotsWithTimeSlot(opts *bind.TransactOpts, slotIds []*big.Int, snapshotterAddresses []common.Address, timeSlots []*big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "assignSnapshotterToSlotsWithTimeSlot", slotIds, snapshotterAddresses, timeSlots)
}

// AssignSnapshotterToSlotsWithTimeSlot is a paid mutator transaction binding the contract method 0xcd6d56bd.
//
// Solidity: function assignSnapshotterToSlotsWithTimeSlot(uint256[] slotIds, address[] snapshotterAddresses, uint256[] timeSlots) returns()
func (_ProtocolState *ProtocolStateSession) AssignSnapshotterToSlotsWithTimeSlot(slotIds []*big.Int, snapshotterAddresses []common.Address, timeSlots []*big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.AssignSnapshotterToSlotsWithTimeSlot(&_ProtocolState.TransactOpts, slotIds, snapshotterAddresses, timeSlots)
}

// AssignSnapshotterToSlotsWithTimeSlot is a paid mutator transaction binding the contract method 0xcd6d56bd.
//
// Solidity: function assignSnapshotterToSlotsWithTimeSlot(uint256[] slotIds, address[] snapshotterAddresses, uint256[] timeSlots) returns()
func (_ProtocolState *ProtocolStateTransactorSession) AssignSnapshotterToSlotsWithTimeSlot(slotIds []*big.Int, snapshotterAddresses []common.Address, timeSlots []*big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.AssignSnapshotterToSlotsWithTimeSlot(&_ProtocolState.TransactOpts, slotIds, snapshotterAddresses, timeSlots)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochId) returns()
func (_ProtocolState *ProtocolStateTransactor) ForceCompleteConsensusSnapshot(opts *bind.TransactOpts, projectId string, epochId *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "forceCompleteConsensusSnapshot", projectId, epochId)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochId) returns()
func (_ProtocolState *ProtocolStateSession) ForceCompleteConsensusSnapshot(projectId string, epochId *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.ForceCompleteConsensusSnapshot(&_ProtocolState.TransactOpts, projectId, epochId)
}

// ForceCompleteConsensusSnapshot is a paid mutator transaction binding the contract method 0x2d5278f3.
//
// Solidity: function forceCompleteConsensusSnapshot(string projectId, uint256 epochId) returns()
func (_ProtocolState *ProtocolStateTransactorSession) ForceCompleteConsensusSnapshot(projectId string, epochId *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.ForceCompleteConsensusSnapshot(&_ProtocolState.TransactOpts, projectId, epochId)
}

// ForceSkipEpoch is a paid mutator transaction binding the contract method 0xf537a3e2.
//
// Solidity: function forceSkipEpoch(uint256 begin, uint256 end) returns()
func (_ProtocolState *ProtocolStateTransactor) ForceSkipEpoch(opts *bind.TransactOpts, begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "forceSkipEpoch", begin, end)
}

// ForceSkipEpoch is a paid mutator transaction binding the contract method 0xf537a3e2.
//
// Solidity: function forceSkipEpoch(uint256 begin, uint256 end) returns()
func (_ProtocolState *ProtocolStateSession) ForceSkipEpoch(begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.ForceSkipEpoch(&_ProtocolState.TransactOpts, begin, end)
}

// ForceSkipEpoch is a paid mutator transaction binding the contract method 0xf537a3e2.
//
// Solidity: function forceSkipEpoch(uint256 begin, uint256 end) returns()
func (_ProtocolState *ProtocolStateTransactorSession) ForceSkipEpoch(begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.ForceSkipEpoch(&_ProtocolState.TransactOpts, begin, end)
}

// ForceStartDay is a paid mutator transaction binding the contract method 0xcd529926.
//
// Solidity: function forceStartDay() returns()
func (_ProtocolState *ProtocolStateTransactor) ForceStartDay(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "forceStartDay")
}

// ForceStartDay is a paid mutator transaction binding the contract method 0xcd529926.
//
// Solidity: function forceStartDay() returns()
func (_ProtocolState *ProtocolStateSession) ForceStartDay() (*types.Transaction, error) {
	return _ProtocolState.Contract.ForceStartDay(&_ProtocolState.TransactOpts)
}

// ForceStartDay is a paid mutator transaction binding the contract method 0xcd529926.
//
// Solidity: function forceStartDay() returns()
func (_ProtocolState *ProtocolStateTransactorSession) ForceStartDay() (*types.Transaction, error) {
	return _ProtocolState.Contract.ForceStartDay(&_ProtocolState.TransactOpts)
}

// LoadCurrentDay is a paid mutator transaction binding the contract method 0x82cdfd43.
//
// Solidity: function loadCurrentDay(uint256 _dayCounter) returns()
func (_ProtocolState *ProtocolStateTransactor) LoadCurrentDay(opts *bind.TransactOpts, _dayCounter *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "loadCurrentDay", _dayCounter)
}

// LoadCurrentDay is a paid mutator transaction binding the contract method 0x82cdfd43.
//
// Solidity: function loadCurrentDay(uint256 _dayCounter) returns()
func (_ProtocolState *ProtocolStateSession) LoadCurrentDay(_dayCounter *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.LoadCurrentDay(&_ProtocolState.TransactOpts, _dayCounter)
}

// LoadCurrentDay is a paid mutator transaction binding the contract method 0x82cdfd43.
//
// Solidity: function loadCurrentDay(uint256 _dayCounter) returns()
func (_ProtocolState *ProtocolStateTransactorSession) LoadCurrentDay(_dayCounter *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.LoadCurrentDay(&_ProtocolState.TransactOpts, _dayCounter)
}

// LoadSlotInfo is a paid mutator transaction binding the contract method 0xa0c6d7f3.
//
// Solidity: function loadSlotInfo(uint256 slotId, address snapshotterAddress, uint256 timeSlot, uint256 rewardPoints, uint256 currentStreak, uint256 currentDaySnapshotCount) returns()
func (_ProtocolState *ProtocolStateTransactor) LoadSlotInfo(opts *bind.TransactOpts, slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int, rewardPoints *big.Int, currentStreak *big.Int, currentDaySnapshotCount *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "loadSlotInfo", slotId, snapshotterAddress, timeSlot, rewardPoints, currentStreak, currentDaySnapshotCount)
}

// LoadSlotInfo is a paid mutator transaction binding the contract method 0xa0c6d7f3.
//
// Solidity: function loadSlotInfo(uint256 slotId, address snapshotterAddress, uint256 timeSlot, uint256 rewardPoints, uint256 currentStreak, uint256 currentDaySnapshotCount) returns()
func (_ProtocolState *ProtocolStateSession) LoadSlotInfo(slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int, rewardPoints *big.Int, currentStreak *big.Int, currentDaySnapshotCount *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.LoadSlotInfo(&_ProtocolState.TransactOpts, slotId, snapshotterAddress, timeSlot, rewardPoints, currentStreak, currentDaySnapshotCount)
}

// LoadSlotInfo is a paid mutator transaction binding the contract method 0xa0c6d7f3.
//
// Solidity: function loadSlotInfo(uint256 slotId, address snapshotterAddress, uint256 timeSlot, uint256 rewardPoints, uint256 currentStreak, uint256 currentDaySnapshotCount) returns()
func (_ProtocolState *ProtocolStateTransactorSession) LoadSlotInfo(slotId *big.Int, snapshotterAddress common.Address, timeSlot *big.Int, rewardPoints *big.Int, currentStreak *big.Int, currentDaySnapshotCount *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.LoadSlotInfo(&_ProtocolState.TransactOpts, slotId, snapshotterAddress, timeSlot, rewardPoints, currentStreak, currentDaySnapshotCount)
}

// LoadSlotSubmissions is a paid mutator transaction binding the contract method 0xbb8ab44b.
//
// Solidity: function loadSlotSubmissions(uint256 slotId, uint256 dayId, uint256 snapshotCount) returns()
func (_ProtocolState *ProtocolStateTransactor) LoadSlotSubmissions(opts *bind.TransactOpts, slotId *big.Int, dayId *big.Int, snapshotCount *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "loadSlotSubmissions", slotId, dayId, snapshotCount)
}

// LoadSlotSubmissions is a paid mutator transaction binding the contract method 0xbb8ab44b.
//
// Solidity: function loadSlotSubmissions(uint256 slotId, uint256 dayId, uint256 snapshotCount) returns()
func (_ProtocolState *ProtocolStateSession) LoadSlotSubmissions(slotId *big.Int, dayId *big.Int, snapshotCount *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.LoadSlotSubmissions(&_ProtocolState.TransactOpts, slotId, dayId, snapshotCount)
}

// LoadSlotSubmissions is a paid mutator transaction binding the contract method 0xbb8ab44b.
//
// Solidity: function loadSlotSubmissions(uint256 slotId, uint256 dayId, uint256 snapshotCount) returns()
func (_ProtocolState *ProtocolStateTransactorSession) LoadSlotSubmissions(slotId *big.Int, dayId *big.Int, snapshotCount *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.LoadSlotSubmissions(&_ProtocolState.TransactOpts, slotId, dayId, snapshotCount)
}

// ReleaseEpoch is a paid mutator transaction binding the contract method 0x132c290f.
//
// Solidity: function releaseEpoch(uint256 begin, uint256 end) returns()
func (_ProtocolState *ProtocolStateTransactor) ReleaseEpoch(opts *bind.TransactOpts, begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "releaseEpoch", begin, end)
}

// ReleaseEpoch is a paid mutator transaction binding the contract method 0x132c290f.
//
// Solidity: function releaseEpoch(uint256 begin, uint256 end) returns()
func (_ProtocolState *ProtocolStateSession) ReleaseEpoch(begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.ReleaseEpoch(&_ProtocolState.TransactOpts, begin, end)
}

// ReleaseEpoch is a paid mutator transaction binding the contract method 0x132c290f.
//
// Solidity: function releaseEpoch(uint256 begin, uint256 end) returns()
func (_ProtocolState *ProtocolStateTransactorSession) ReleaseEpoch(begin *big.Int, end *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.ReleaseEpoch(&_ProtocolState.TransactOpts, begin, end)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProtocolState *ProtocolStateTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProtocolState *ProtocolStateSession) RenounceOwnership() (*types.Transaction, error) {
	return _ProtocolState.Contract.RenounceOwnership(&_ProtocolState.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProtocolState *ProtocolStateTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ProtocolState.Contract.RenounceOwnership(&_ProtocolState.TransactOpts)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0x0fbe203a.
//
// Solidity: function submitSnapshot(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature) returns()
func (_ProtocolState *ProtocolStateTransactor) SubmitSnapshot(opts *bind.TransactOpts, slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "submitSnapshot", slotId, snapshotCid, epochId, projectId, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0x0fbe203a.
//
// Solidity: function submitSnapshot(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature) returns()
func (_ProtocolState *ProtocolStateSession) SubmitSnapshot(slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ProtocolState.Contract.SubmitSnapshot(&_ProtocolState.TransactOpts, slotId, snapshotCid, epochId, projectId, request, signature)
}

// SubmitSnapshot is a paid mutator transaction binding the contract method 0x0fbe203a.
//
// Solidity: function submitSnapshot(uint256 slotId, string snapshotCid, uint256 epochId, string projectId, (uint256,uint256,string,uint256,string) request, bytes signature) returns()
func (_ProtocolState *ProtocolStateTransactorSession) SubmitSnapshot(slotId *big.Int, snapshotCid string, epochId *big.Int, projectId string, request PowerloomProtocolStateRequest, signature []byte) (*types.Transaction, error) {
	return _ProtocolState.Contract.SubmitSnapshot(&_ProtocolState.TransactOpts, slotId, snapshotCid, epochId, projectId, request, signature)
}

// ToggleRewards is a paid mutator transaction binding the contract method 0x95268408.
//
// Solidity: function toggleRewards() returns()
func (_ProtocolState *ProtocolStateTransactor) ToggleRewards(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "toggleRewards")
}

// ToggleRewards is a paid mutator transaction binding the contract method 0x95268408.
//
// Solidity: function toggleRewards() returns()
func (_ProtocolState *ProtocolStateSession) ToggleRewards() (*types.Transaction, error) {
	return _ProtocolState.Contract.ToggleRewards(&_ProtocolState.TransactOpts)
}

// ToggleRewards is a paid mutator transaction binding the contract method 0x95268408.
//
// Solidity: function toggleRewards() returns()
func (_ProtocolState *ProtocolStateTransactorSession) ToggleRewards() (*types.Transaction, error) {
	return _ProtocolState.Contract.ToggleRewards(&_ProtocolState.TransactOpts)
}

// ToggleTimeSlotCheck is a paid mutator transaction binding the contract method 0x75f2d951.
//
// Solidity: function toggleTimeSlotCheck() returns()
func (_ProtocolState *ProtocolStateTransactor) ToggleTimeSlotCheck(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "toggleTimeSlotCheck")
}

// ToggleTimeSlotCheck is a paid mutator transaction binding the contract method 0x75f2d951.
//
// Solidity: function toggleTimeSlotCheck() returns()
func (_ProtocolState *ProtocolStateSession) ToggleTimeSlotCheck() (*types.Transaction, error) {
	return _ProtocolState.Contract.ToggleTimeSlotCheck(&_ProtocolState.TransactOpts)
}

// ToggleTimeSlotCheck is a paid mutator transaction binding the contract method 0x75f2d951.
//
// Solidity: function toggleTimeSlotCheck() returns()
func (_ProtocolState *ProtocolStateTransactorSession) ToggleTimeSlotCheck() (*types.Transaction, error) {
	return _ProtocolState.Contract.ToggleTimeSlotCheck(&_ProtocolState.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ProtocolState *ProtocolStateTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ProtocolState *ProtocolStateSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ProtocolState.Contract.TransferOwnership(&_ProtocolState.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ProtocolState *ProtocolStateTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ProtocolState.Contract.TransferOwnership(&_ProtocolState.TransactOpts, newOwner)
}

// UpdateAllowedProjectType is a paid mutator transaction binding the contract method 0x562209f3.
//
// Solidity: function updateAllowedProjectType(string _projectType, bool _status) returns()
func (_ProtocolState *ProtocolStateTransactor) UpdateAllowedProjectType(opts *bind.TransactOpts, _projectType string, _status bool) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "updateAllowedProjectType", _projectType, _status)
}

// UpdateAllowedProjectType is a paid mutator transaction binding the contract method 0x562209f3.
//
// Solidity: function updateAllowedProjectType(string _projectType, bool _status) returns()
func (_ProtocolState *ProtocolStateSession) UpdateAllowedProjectType(_projectType string, _status bool) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateAllowedProjectType(&_ProtocolState.TransactOpts, _projectType, _status)
}

// UpdateAllowedProjectType is a paid mutator transaction binding the contract method 0x562209f3.
//
// Solidity: function updateAllowedProjectType(string _projectType, bool _status) returns()
func (_ProtocolState *ProtocolStateTransactorSession) UpdateAllowedProjectType(_projectType string, _status bool) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateAllowedProjectType(&_ProtocolState.TransactOpts, _projectType, _status)
}

// UpdateDailySnapshotQuota is a paid mutator transaction binding the contract method 0x6d26e630.
//
// Solidity: function updateDailySnapshotQuota(uint256 dailySnapshotQuota) returns()
func (_ProtocolState *ProtocolStateTransactor) UpdateDailySnapshotQuota(opts *bind.TransactOpts, dailySnapshotQuota *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "updateDailySnapshotQuota", dailySnapshotQuota)
}

// UpdateDailySnapshotQuota is a paid mutator transaction binding the contract method 0x6d26e630.
//
// Solidity: function updateDailySnapshotQuota(uint256 dailySnapshotQuota) returns()
func (_ProtocolState *ProtocolStateSession) UpdateDailySnapshotQuota(dailySnapshotQuota *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateDailySnapshotQuota(&_ProtocolState.TransactOpts, dailySnapshotQuota)
}

// UpdateDailySnapshotQuota is a paid mutator transaction binding the contract method 0x6d26e630.
//
// Solidity: function updateDailySnapshotQuota(uint256 dailySnapshotQuota) returns()
func (_ProtocolState *ProtocolStateTransactorSession) UpdateDailySnapshotQuota(dailySnapshotQuota *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateDailySnapshotQuota(&_ProtocolState.TransactOpts, dailySnapshotQuota)
}

// UpdateMasterSnapshotters is a paid mutator transaction binding the contract method 0x8f9eacf3.
//
// Solidity: function updateMasterSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ProtocolState *ProtocolStateTransactor) UpdateMasterSnapshotters(opts *bind.TransactOpts, _snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "updateMasterSnapshotters", _snapshotters, _status)
}

// UpdateMasterSnapshotters is a paid mutator transaction binding the contract method 0x8f9eacf3.
//
// Solidity: function updateMasterSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ProtocolState *ProtocolStateSession) UpdateMasterSnapshotters(_snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateMasterSnapshotters(&_ProtocolState.TransactOpts, _snapshotters, _status)
}

// UpdateMasterSnapshotters is a paid mutator transaction binding the contract method 0x8f9eacf3.
//
// Solidity: function updateMasterSnapshotters(address[] _snapshotters, bool[] _status) returns()
func (_ProtocolState *ProtocolStateTransactorSession) UpdateMasterSnapshotters(_snapshotters []common.Address, _status []bool) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateMasterSnapshotters(&_ProtocolState.TransactOpts, _snapshotters, _status)
}

// UpdateMinSnapshottersForConsensus is a paid mutator transaction binding the contract method 0x38deacd3.
//
// Solidity: function updateMinSnapshottersForConsensus(uint256 _minSubmissionsForConsensus) returns()
func (_ProtocolState *ProtocolStateTransactor) UpdateMinSnapshottersForConsensus(opts *bind.TransactOpts, _minSubmissionsForConsensus *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "updateMinSnapshottersForConsensus", _minSubmissionsForConsensus)
}

// UpdateMinSnapshottersForConsensus is a paid mutator transaction binding the contract method 0x38deacd3.
//
// Solidity: function updateMinSnapshottersForConsensus(uint256 _minSubmissionsForConsensus) returns()
func (_ProtocolState *ProtocolStateSession) UpdateMinSnapshottersForConsensus(_minSubmissionsForConsensus *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateMinSnapshottersForConsensus(&_ProtocolState.TransactOpts, _minSubmissionsForConsensus)
}

// UpdateMinSnapshottersForConsensus is a paid mutator transaction binding the contract method 0x38deacd3.
//
// Solidity: function updateMinSnapshottersForConsensus(uint256 _minSubmissionsForConsensus) returns()
func (_ProtocolState *ProtocolStateTransactorSession) UpdateMinSnapshottersForConsensus(_minSubmissionsForConsensus *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateMinSnapshottersForConsensus(&_ProtocolState.TransactOpts, _minSubmissionsForConsensus)
}

// UpdateSnapshotSubmissionWindow is a paid mutator transaction binding the contract method 0x9b2f89ce.
//
// Solidity: function updateSnapshotSubmissionWindow(uint256 newsnapshotSubmissionWindow) returns()
func (_ProtocolState *ProtocolStateTransactor) UpdateSnapshotSubmissionWindow(opts *bind.TransactOpts, newsnapshotSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "updateSnapshotSubmissionWindow", newsnapshotSubmissionWindow)
}

// UpdateSnapshotSubmissionWindow is a paid mutator transaction binding the contract method 0x9b2f89ce.
//
// Solidity: function updateSnapshotSubmissionWindow(uint256 newsnapshotSubmissionWindow) returns()
func (_ProtocolState *ProtocolStateSession) UpdateSnapshotSubmissionWindow(newsnapshotSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateSnapshotSubmissionWindow(&_ProtocolState.TransactOpts, newsnapshotSubmissionWindow)
}

// UpdateSnapshotSubmissionWindow is a paid mutator transaction binding the contract method 0x9b2f89ce.
//
// Solidity: function updateSnapshotSubmissionWindow(uint256 newsnapshotSubmissionWindow) returns()
func (_ProtocolState *ProtocolStateTransactorSession) UpdateSnapshotSubmissionWindow(newsnapshotSubmissionWindow *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateSnapshotSubmissionWindow(&_ProtocolState.TransactOpts, newsnapshotSubmissionWindow)
}

// UpdateValidators is a paid mutator transaction binding the contract method 0x1b0b3ae3.
//
// Solidity: function updateValidators(address[] _validators, bool[] _status) returns()
func (_ProtocolState *ProtocolStateTransactor) UpdateValidators(opts *bind.TransactOpts, _validators []common.Address, _status []bool) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "updateValidators", _validators, _status)
}

// UpdateValidators is a paid mutator transaction binding the contract method 0x1b0b3ae3.
//
// Solidity: function updateValidators(address[] _validators, bool[] _status) returns()
func (_ProtocolState *ProtocolStateSession) UpdateValidators(_validators []common.Address, _status []bool) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateValidators(&_ProtocolState.TransactOpts, _validators, _status)
}

// UpdateValidators is a paid mutator transaction binding the contract method 0x1b0b3ae3.
//
// Solidity: function updateValidators(address[] _validators, bool[] _status) returns()
func (_ProtocolState *ProtocolStateTransactorSession) UpdateValidators(_validators []common.Address, _status []bool) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdateValidators(&_ProtocolState.TransactOpts, _validators, _status)
}

// UpdaterewardBasePoints is a paid mutator transaction binding the contract method 0x801939d7.
//
// Solidity: function updaterewardBasePoints(uint256 newrewardBasePoints) returns()
func (_ProtocolState *ProtocolStateTransactor) UpdaterewardBasePoints(opts *bind.TransactOpts, newrewardBasePoints *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "updaterewardBasePoints", newrewardBasePoints)
}

// UpdaterewardBasePoints is a paid mutator transaction binding the contract method 0x801939d7.
//
// Solidity: function updaterewardBasePoints(uint256 newrewardBasePoints) returns()
func (_ProtocolState *ProtocolStateSession) UpdaterewardBasePoints(newrewardBasePoints *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdaterewardBasePoints(&_ProtocolState.TransactOpts, newrewardBasePoints)
}

// UpdaterewardBasePoints is a paid mutator transaction binding the contract method 0x801939d7.
//
// Solidity: function updaterewardBasePoints(uint256 newrewardBasePoints) returns()
func (_ProtocolState *ProtocolStateTransactorSession) UpdaterewardBasePoints(newrewardBasePoints *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdaterewardBasePoints(&_ProtocolState.TransactOpts, newrewardBasePoints)
}

// UpdatestreakBonusPoints is a paid mutator transaction binding the contract method 0xa8939fa5.
//
// Solidity: function updatestreakBonusPoints(uint256 newstreakBonusPoints) returns()
func (_ProtocolState *ProtocolStateTransactor) UpdatestreakBonusPoints(opts *bind.TransactOpts, newstreakBonusPoints *big.Int) (*types.Transaction, error) {
	return _ProtocolState.contract.Transact(opts, "updatestreakBonusPoints", newstreakBonusPoints)
}

// UpdatestreakBonusPoints is a paid mutator transaction binding the contract method 0xa8939fa5.
//
// Solidity: function updatestreakBonusPoints(uint256 newstreakBonusPoints) returns()
func (_ProtocolState *ProtocolStateSession) UpdatestreakBonusPoints(newstreakBonusPoints *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdatestreakBonusPoints(&_ProtocolState.TransactOpts, newstreakBonusPoints)
}

// UpdatestreakBonusPoints is a paid mutator transaction binding the contract method 0xa8939fa5.
//
// Solidity: function updatestreakBonusPoints(uint256 newstreakBonusPoints) returns()
func (_ProtocolState *ProtocolStateTransactorSession) UpdatestreakBonusPoints(newstreakBonusPoints *big.Int) (*types.Transaction, error) {
	return _ProtocolState.Contract.UpdatestreakBonusPoints(&_ProtocolState.TransactOpts, newstreakBonusPoints)
}

// ProtocolStateDailyTaskCompletedEventIterator is returned from FilterDailyTaskCompletedEvent and is used to iterate over the raw logs and unpacked data for DailyTaskCompletedEvent events raised by the ProtocolState contract.
type ProtocolStateDailyTaskCompletedEventIterator struct {
	Event *ProtocolStateDailyTaskCompletedEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateDailyTaskCompletedEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateDailyTaskCompletedEvent)
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
		it.Event = new(ProtocolStateDailyTaskCompletedEvent)
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
func (it *ProtocolStateDailyTaskCompletedEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateDailyTaskCompletedEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateDailyTaskCompletedEvent represents a DailyTaskCompletedEvent event raised by the ProtocolState contract.
type ProtocolStateDailyTaskCompletedEvent struct {
	SnapshotterAddress common.Address
	SlotId             *big.Int
	DayId              *big.Int
	Timestamp          *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterDailyTaskCompletedEvent is a free log retrieval operation binding the contract event 0x34c900c4105cef3bd58c4b7d2b6fe54f1f64845d5bd5ed2e2e92b52aed2d58ae.
//
// Solidity: event DailyTaskCompletedEvent(address snapshotterAddress, uint256 slotId, uint256 dayId, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) FilterDailyTaskCompletedEvent(opts *bind.FilterOpts) (*ProtocolStateDailyTaskCompletedEventIterator, error) {

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "DailyTaskCompletedEvent")
	if err != nil {
		return nil, err
	}
	return &ProtocolStateDailyTaskCompletedEventIterator{contract: _ProtocolState.contract, event: "DailyTaskCompletedEvent", logs: logs, sub: sub}, nil
}

// WatchDailyTaskCompletedEvent is a free log subscription operation binding the contract event 0x34c900c4105cef3bd58c4b7d2b6fe54f1f64845d5bd5ed2e2e92b52aed2d58ae.
//
// Solidity: event DailyTaskCompletedEvent(address snapshotterAddress, uint256 slotId, uint256 dayId, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) WatchDailyTaskCompletedEvent(opts *bind.WatchOpts, sink chan<- *ProtocolStateDailyTaskCompletedEvent) (event.Subscription, error) {

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "DailyTaskCompletedEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateDailyTaskCompletedEvent)
				if err := _ProtocolState.contract.UnpackLog(event, "DailyTaskCompletedEvent", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseDailyTaskCompletedEvent(log types.Log) (*ProtocolStateDailyTaskCompletedEvent, error) {
	event := new(ProtocolStateDailyTaskCompletedEvent)
	if err := _ProtocolState.contract.UnpackLog(event, "DailyTaskCompletedEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateDayStartedEventIterator is returned from FilterDayStartedEvent and is used to iterate over the raw logs and unpacked data for DayStartedEvent events raised by the ProtocolState contract.
type ProtocolStateDayStartedEventIterator struct {
	Event *ProtocolStateDayStartedEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateDayStartedEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateDayStartedEvent)
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
		it.Event = new(ProtocolStateDayStartedEvent)
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
func (it *ProtocolStateDayStartedEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateDayStartedEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateDayStartedEvent represents a DayStartedEvent event raised by the ProtocolState contract.
type ProtocolStateDayStartedEvent struct {
	DayId     *big.Int
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDayStartedEvent is a free log retrieval operation binding the contract event 0xf391963fbbcec4cbb1f4a6c915c531364db26c103d31434223c3bddb703c94fe.
//
// Solidity: event DayStartedEvent(uint256 dayId, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) FilterDayStartedEvent(opts *bind.FilterOpts) (*ProtocolStateDayStartedEventIterator, error) {

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "DayStartedEvent")
	if err != nil {
		return nil, err
	}
	return &ProtocolStateDayStartedEventIterator{contract: _ProtocolState.contract, event: "DayStartedEvent", logs: logs, sub: sub}, nil
}

// WatchDayStartedEvent is a free log subscription operation binding the contract event 0xf391963fbbcec4cbb1f4a6c915c531364db26c103d31434223c3bddb703c94fe.
//
// Solidity: event DayStartedEvent(uint256 dayId, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) WatchDayStartedEvent(opts *bind.WatchOpts, sink chan<- *ProtocolStateDayStartedEvent) (event.Subscription, error) {

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "DayStartedEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateDayStartedEvent)
				if err := _ProtocolState.contract.UnpackLog(event, "DayStartedEvent", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseDayStartedEvent(log types.Log) (*ProtocolStateDayStartedEvent, error) {
	event := new(ProtocolStateDayStartedEvent)
	if err := _ProtocolState.contract.UnpackLog(event, "DayStartedEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateDelayedSnapshotSubmittedIterator is returned from FilterDelayedSnapshotSubmitted and is used to iterate over the raw logs and unpacked data for DelayedSnapshotSubmitted events raised by the ProtocolState contract.
type ProtocolStateDelayedSnapshotSubmittedIterator struct {
	Event *ProtocolStateDelayedSnapshotSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateDelayedSnapshotSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateDelayedSnapshotSubmitted)
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
		it.Event = new(ProtocolStateDelayedSnapshotSubmitted)
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
func (it *ProtocolStateDelayedSnapshotSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateDelayedSnapshotSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateDelayedSnapshotSubmitted represents a DelayedSnapshotSubmitted event raised by the ProtocolState contract.
type ProtocolStateDelayedSnapshotSubmitted struct {
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
func (_ProtocolState *ProtocolStateFilterer) FilterDelayedSnapshotSubmitted(opts *bind.FilterOpts, snapshotterAddr []common.Address, epochId []*big.Int) (*ProtocolStateDelayedSnapshotSubmittedIterator, error) {

	var snapshotterAddrRule []interface{}
	for _, snapshotterAddrItem := range snapshotterAddr {
		snapshotterAddrRule = append(snapshotterAddrRule, snapshotterAddrItem)
	}

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "DelayedSnapshotSubmitted", snapshotterAddrRule, epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ProtocolStateDelayedSnapshotSubmittedIterator{contract: _ProtocolState.contract, event: "DelayedSnapshotSubmitted", logs: logs, sub: sub}, nil
}

// WatchDelayedSnapshotSubmitted is a free log subscription operation binding the contract event 0x8eb09d3de43d012d958db78fcc692d37e9f9c0cc2af1cdd8c58617832af17d31.
//
// Solidity: event DelayedSnapshotSubmitted(address indexed snapshotterAddr, uint256 slotId, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) WatchDelayedSnapshotSubmitted(opts *bind.WatchOpts, sink chan<- *ProtocolStateDelayedSnapshotSubmitted, snapshotterAddr []common.Address, epochId []*big.Int) (event.Subscription, error) {

	var snapshotterAddrRule []interface{}
	for _, snapshotterAddrItem := range snapshotterAddr {
		snapshotterAddrRule = append(snapshotterAddrRule, snapshotterAddrItem)
	}

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "DelayedSnapshotSubmitted", snapshotterAddrRule, epochIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateDelayedSnapshotSubmitted)
				if err := _ProtocolState.contract.UnpackLog(event, "DelayedSnapshotSubmitted", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseDelayedSnapshotSubmitted(log types.Log) (*ProtocolStateDelayedSnapshotSubmitted, error) {
	event := new(ProtocolStateDelayedSnapshotSubmitted)
	if err := _ProtocolState.contract.UnpackLog(event, "DelayedSnapshotSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the ProtocolState contract.
type ProtocolStateEIP712DomainChangedIterator struct {
	Event *ProtocolStateEIP712DomainChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateEIP712DomainChanged)
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
		it.Event = new(ProtocolStateEIP712DomainChanged)
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
func (it *ProtocolStateEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateEIP712DomainChanged represents a EIP712DomainChanged event raised by the ProtocolState contract.
type ProtocolStateEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_ProtocolState *ProtocolStateFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*ProtocolStateEIP712DomainChangedIterator, error) {

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &ProtocolStateEIP712DomainChangedIterator{contract: _ProtocolState.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_ProtocolState *ProtocolStateFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *ProtocolStateEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateEIP712DomainChanged)
				if err := _ProtocolState.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseEIP712DomainChanged(log types.Log) (*ProtocolStateEIP712DomainChanged, error) {
	event := new(ProtocolStateEIP712DomainChanged)
	if err := _ProtocolState.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateEpochReleasedIterator is returned from FilterEpochReleased and is used to iterate over the raw logs and unpacked data for EpochReleased events raised by the ProtocolState contract.
type ProtocolStateEpochReleasedIterator struct {
	Event *ProtocolStateEpochReleased // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateEpochReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateEpochReleased)
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
		it.Event = new(ProtocolStateEpochReleased)
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
func (it *ProtocolStateEpochReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateEpochReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateEpochReleased represents a EpochReleased event raised by the ProtocolState contract.
type ProtocolStateEpochReleased struct {
	EpochId   *big.Int
	Begin     *big.Int
	End       *big.Int
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEpochReleased is a free log retrieval operation binding the contract event 0x108f87075a74f81fa2271fdf9fc0883a1811431182601fc65d24513970336640.
//
// Solidity: event EpochReleased(uint256 indexed epochId, uint256 begin, uint256 end, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) FilterEpochReleased(opts *bind.FilterOpts, epochId []*big.Int) (*ProtocolStateEpochReleasedIterator, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "EpochReleased", epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ProtocolStateEpochReleasedIterator{contract: _ProtocolState.contract, event: "EpochReleased", logs: logs, sub: sub}, nil
}

// WatchEpochReleased is a free log subscription operation binding the contract event 0x108f87075a74f81fa2271fdf9fc0883a1811431182601fc65d24513970336640.
//
// Solidity: event EpochReleased(uint256 indexed epochId, uint256 begin, uint256 end, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) WatchEpochReleased(opts *bind.WatchOpts, sink chan<- *ProtocolStateEpochReleased, epochId []*big.Int) (event.Subscription, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "EpochReleased", epochIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateEpochReleased)
				if err := _ProtocolState.contract.UnpackLog(event, "EpochReleased", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseEpochReleased(log types.Log) (*ProtocolStateEpochReleased, error) {
	event := new(ProtocolStateEpochReleased)
	if err := _ProtocolState.contract.UnpackLog(event, "EpochReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ProtocolState contract.
type ProtocolStateOwnershipTransferredIterator struct {
	Event *ProtocolStateOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateOwnershipTransferred)
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
		it.Event = new(ProtocolStateOwnershipTransferred)
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
func (it *ProtocolStateOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateOwnershipTransferred represents a OwnershipTransferred event raised by the ProtocolState contract.
type ProtocolStateOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ProtocolState *ProtocolStateFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ProtocolStateOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ProtocolStateOwnershipTransferredIterator{contract: _ProtocolState.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ProtocolState *ProtocolStateFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ProtocolStateOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateOwnershipTransferred)
				if err := _ProtocolState.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseOwnershipTransferred(log types.Log) (*ProtocolStateOwnershipTransferred, error) {
	event := new(ProtocolStateOwnershipTransferred)
	if err := _ProtocolState.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateProjectTypeUpdatedIterator is returned from FilterProjectTypeUpdated and is used to iterate over the raw logs and unpacked data for ProjectTypeUpdated events raised by the ProtocolState contract.
type ProtocolStateProjectTypeUpdatedIterator struct {
	Event *ProtocolStateProjectTypeUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateProjectTypeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateProjectTypeUpdated)
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
		it.Event = new(ProtocolStateProjectTypeUpdated)
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
func (it *ProtocolStateProjectTypeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateProjectTypeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateProjectTypeUpdated represents a ProjectTypeUpdated event raised by the ProtocolState contract.
type ProtocolStateProjectTypeUpdated struct {
	ProjectType   string
	Allowed       bool
	EnableEpochId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterProjectTypeUpdated is a free log retrieval operation binding the contract event 0xcd1b466e0de4e12ecf0bc286450d3e5b4aa88db70272ed5c11508b16acb35bfc.
//
// Solidity: event ProjectTypeUpdated(string projectType, bool allowed, uint256 enableEpochId)
func (_ProtocolState *ProtocolStateFilterer) FilterProjectTypeUpdated(opts *bind.FilterOpts) (*ProtocolStateProjectTypeUpdatedIterator, error) {

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "ProjectTypeUpdated")
	if err != nil {
		return nil, err
	}
	return &ProtocolStateProjectTypeUpdatedIterator{contract: _ProtocolState.contract, event: "ProjectTypeUpdated", logs: logs, sub: sub}, nil
}

// WatchProjectTypeUpdated is a free log subscription operation binding the contract event 0xcd1b466e0de4e12ecf0bc286450d3e5b4aa88db70272ed5c11508b16acb35bfc.
//
// Solidity: event ProjectTypeUpdated(string projectType, bool allowed, uint256 enableEpochId)
func (_ProtocolState *ProtocolStateFilterer) WatchProjectTypeUpdated(opts *bind.WatchOpts, sink chan<- *ProtocolStateProjectTypeUpdated) (event.Subscription, error) {

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "ProjectTypeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateProjectTypeUpdated)
				if err := _ProtocolState.contract.UnpackLog(event, "ProjectTypeUpdated", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseProjectTypeUpdated(log types.Log) (*ProtocolStateProjectTypeUpdated, error) {
	event := new(ProtocolStateProjectTypeUpdated)
	if err := _ProtocolState.contract.UnpackLog(event, "ProjectTypeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateSnapshotFinalizedIterator is returned from FilterSnapshotFinalized and is used to iterate over the raw logs and unpacked data for SnapshotFinalized events raised by the ProtocolState contract.
type ProtocolStateSnapshotFinalizedIterator struct {
	Event *ProtocolStateSnapshotFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateSnapshotFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateSnapshotFinalized)
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
		it.Event = new(ProtocolStateSnapshotFinalized)
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
func (it *ProtocolStateSnapshotFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateSnapshotFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateSnapshotFinalized represents a SnapshotFinalized event raised by the ProtocolState contract.
type ProtocolStateSnapshotFinalized struct {
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
func (_ProtocolState *ProtocolStateFilterer) FilterSnapshotFinalized(opts *bind.FilterOpts, epochId []*big.Int) (*ProtocolStateSnapshotFinalizedIterator, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "SnapshotFinalized", epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ProtocolStateSnapshotFinalizedIterator{contract: _ProtocolState.contract, event: "SnapshotFinalized", logs: logs, sub: sub}, nil
}

// WatchSnapshotFinalized is a free log subscription operation binding the contract event 0x029bad7ce411d223096bf56c17c32a3659aa2703de0161497e6284e7fc288c35.
//
// Solidity: event SnapshotFinalized(uint256 indexed epochId, uint256 epochEnd, string projectId, string snapshotCid, uint256 finalizedSnapshotCount, uint256 totalReceivedCount, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) WatchSnapshotFinalized(opts *bind.WatchOpts, sink chan<- *ProtocolStateSnapshotFinalized, epochId []*big.Int) (event.Subscription, error) {

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "SnapshotFinalized", epochIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateSnapshotFinalized)
				if err := _ProtocolState.contract.UnpackLog(event, "SnapshotFinalized", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseSnapshotFinalized(log types.Log) (*ProtocolStateSnapshotFinalized, error) {
	event := new(ProtocolStateSnapshotFinalized)
	if err := _ProtocolState.contract.UnpackLog(event, "SnapshotFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateSnapshotSubmittedIterator is returned from FilterSnapshotSubmitted and is used to iterate over the raw logs and unpacked data for SnapshotSubmitted events raised by the ProtocolState contract.
type ProtocolStateSnapshotSubmittedIterator struct {
	Event *ProtocolStateSnapshotSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateSnapshotSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateSnapshotSubmitted)
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
		it.Event = new(ProtocolStateSnapshotSubmitted)
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
func (it *ProtocolStateSnapshotSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateSnapshotSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateSnapshotSubmitted represents a SnapshotSubmitted event raised by the ProtocolState contract.
type ProtocolStateSnapshotSubmitted struct {
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
func (_ProtocolState *ProtocolStateFilterer) FilterSnapshotSubmitted(opts *bind.FilterOpts, snapshotterAddr []common.Address, epochId []*big.Int) (*ProtocolStateSnapshotSubmittedIterator, error) {

	var snapshotterAddrRule []interface{}
	for _, snapshotterAddrItem := range snapshotterAddr {
		snapshotterAddrRule = append(snapshotterAddrRule, snapshotterAddrItem)
	}

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "SnapshotSubmitted", snapshotterAddrRule, epochIdRule)
	if err != nil {
		return nil, err
	}
	return &ProtocolStateSnapshotSubmittedIterator{contract: _ProtocolState.contract, event: "SnapshotSubmitted", logs: logs, sub: sub}, nil
}

// WatchSnapshotSubmitted is a free log subscription operation binding the contract event 0x8d9341fa1766ade9f55ddb87e37a11648afefce0f76a389675dd56a5d555b8d3.
//
// Solidity: event SnapshotSubmitted(address indexed snapshotterAddr, uint256 slotId, string snapshotCid, uint256 indexed epochId, string projectId, uint256 timestamp)
func (_ProtocolState *ProtocolStateFilterer) WatchSnapshotSubmitted(opts *bind.WatchOpts, sink chan<- *ProtocolStateSnapshotSubmitted, snapshotterAddr []common.Address, epochId []*big.Int) (event.Subscription, error) {

	var snapshotterAddrRule []interface{}
	for _, snapshotterAddrItem := range snapshotterAddr {
		snapshotterAddrRule = append(snapshotterAddrRule, snapshotterAddrItem)
	}

	var epochIdRule []interface{}
	for _, epochIdItem := range epochId {
		epochIdRule = append(epochIdRule, epochIdItem)
	}

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "SnapshotSubmitted", snapshotterAddrRule, epochIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateSnapshotSubmitted)
				if err := _ProtocolState.contract.UnpackLog(event, "SnapshotSubmitted", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseSnapshotSubmitted(log types.Log) (*ProtocolStateSnapshotSubmitted, error) {
	event := new(ProtocolStateSnapshotSubmitted)
	if err := _ProtocolState.contract.UnpackLog(event, "SnapshotSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateValidatorsUpdatedIterator is returned from FilterValidatorsUpdated and is used to iterate over the raw logs and unpacked data for ValidatorsUpdated events raised by the ProtocolState contract.
type ProtocolStateValidatorsUpdatedIterator struct {
	Event *ProtocolStateValidatorsUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateValidatorsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateValidatorsUpdated)
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
		it.Event = new(ProtocolStateValidatorsUpdated)
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
func (it *ProtocolStateValidatorsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateValidatorsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateValidatorsUpdated represents a ValidatorsUpdated event raised by the ProtocolState contract.
type ProtocolStateValidatorsUpdated struct {
	ValidatorAddress common.Address
	Allowed          bool
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterValidatorsUpdated is a free log retrieval operation binding the contract event 0x7f3079c058f3e3dee87048158309898b46e9741ff53b6c7a3afac7c370649afc.
//
// Solidity: event ValidatorsUpdated(address validatorAddress, bool allowed)
func (_ProtocolState *ProtocolStateFilterer) FilterValidatorsUpdated(opts *bind.FilterOpts) (*ProtocolStateValidatorsUpdatedIterator, error) {

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "ValidatorsUpdated")
	if err != nil {
		return nil, err
	}
	return &ProtocolStateValidatorsUpdatedIterator{contract: _ProtocolState.contract, event: "ValidatorsUpdated", logs: logs, sub: sub}, nil
}

// WatchValidatorsUpdated is a free log subscription operation binding the contract event 0x7f3079c058f3e3dee87048158309898b46e9741ff53b6c7a3afac7c370649afc.
//
// Solidity: event ValidatorsUpdated(address validatorAddress, bool allowed)
func (_ProtocolState *ProtocolStateFilterer) WatchValidatorsUpdated(opts *bind.WatchOpts, sink chan<- *ProtocolStateValidatorsUpdated) (event.Subscription, error) {

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "ValidatorsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateValidatorsUpdated)
				if err := _ProtocolState.contract.UnpackLog(event, "ValidatorsUpdated", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseValidatorsUpdated(log types.Log) (*ProtocolStateValidatorsUpdated, error) {
	event := new(ProtocolStateValidatorsUpdated)
	if err := _ProtocolState.contract.UnpackLog(event, "ValidatorsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateAllSnapshottersUpdatedIterator is returned from FilterAllSnapshottersUpdated and is used to iterate over the raw logs and unpacked data for AllSnapshottersUpdated events raised by the ProtocolState contract.
type ProtocolStateAllSnapshottersUpdatedIterator struct {
	Event *ProtocolStateAllSnapshottersUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateAllSnapshottersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateAllSnapshottersUpdated)
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
		it.Event = new(ProtocolStateAllSnapshottersUpdated)
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
func (it *ProtocolStateAllSnapshottersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateAllSnapshottersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateAllSnapshottersUpdated represents a AllSnapshottersUpdated event raised by the ProtocolState contract.
type ProtocolStateAllSnapshottersUpdated struct {
	SnapshotterAddress common.Address
	Allowed            bool
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterAllSnapshottersUpdated is a free log retrieval operation binding the contract event 0x743e47fbcd2e3a64a2ab8f5dcdeb4c17f892c17c9f6ab58d3a5c235953d60058.
//
// Solidity: event allSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ProtocolState *ProtocolStateFilterer) FilterAllSnapshottersUpdated(opts *bind.FilterOpts) (*ProtocolStateAllSnapshottersUpdatedIterator, error) {

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "allSnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return &ProtocolStateAllSnapshottersUpdatedIterator{contract: _ProtocolState.contract, event: "allSnapshottersUpdated", logs: logs, sub: sub}, nil
}

// WatchAllSnapshottersUpdated is a free log subscription operation binding the contract event 0x743e47fbcd2e3a64a2ab8f5dcdeb4c17f892c17c9f6ab58d3a5c235953d60058.
//
// Solidity: event allSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ProtocolState *ProtocolStateFilterer) WatchAllSnapshottersUpdated(opts *bind.WatchOpts, sink chan<- *ProtocolStateAllSnapshottersUpdated) (event.Subscription, error) {

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "allSnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateAllSnapshottersUpdated)
				if err := _ProtocolState.contract.UnpackLog(event, "allSnapshottersUpdated", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseAllSnapshottersUpdated(log types.Log) (*ProtocolStateAllSnapshottersUpdated, error) {
	event := new(ProtocolStateAllSnapshottersUpdated)
	if err := _ProtocolState.contract.UnpackLog(event, "allSnapshottersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProtocolStateMasterSnapshottersUpdatedIterator is returned from FilterMasterSnapshottersUpdated and is used to iterate over the raw logs and unpacked data for MasterSnapshottersUpdated events raised by the ProtocolState contract.
type ProtocolStateMasterSnapshottersUpdatedIterator struct {
	Event *ProtocolStateMasterSnapshottersUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log          // Log channel receiving the found contract events
	sub  interfaces.Subscription // Subscription for errors, completion and termination
	done bool                    // Whether the subscription completed delivering logs
	fail error                   // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProtocolStateMasterSnapshottersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProtocolStateMasterSnapshottersUpdated)
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
		it.Event = new(ProtocolStateMasterSnapshottersUpdated)
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
func (it *ProtocolStateMasterSnapshottersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProtocolStateMasterSnapshottersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProtocolStateMasterSnapshottersUpdated represents a MasterSnapshottersUpdated event raised by the ProtocolState contract.
type ProtocolStateMasterSnapshottersUpdated struct {
	SnapshotterAddress common.Address
	Allowed            bool
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterMasterSnapshottersUpdated is a free log retrieval operation binding the contract event 0x9c2d4c2b4cf1ca90e31b1448712dae908d3568785b68a411f6617479c2b9913b.
//
// Solidity: event masterSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ProtocolState *ProtocolStateFilterer) FilterMasterSnapshottersUpdated(opts *bind.FilterOpts) (*ProtocolStateMasterSnapshottersUpdatedIterator, error) {

	logs, sub, err := _ProtocolState.contract.FilterLogs(opts, "masterSnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return &ProtocolStateMasterSnapshottersUpdatedIterator{contract: _ProtocolState.contract, event: "masterSnapshottersUpdated", logs: logs, sub: sub}, nil
}

// WatchMasterSnapshottersUpdated is a free log subscription operation binding the contract event 0x9c2d4c2b4cf1ca90e31b1448712dae908d3568785b68a411f6617479c2b9913b.
//
// Solidity: event masterSnapshottersUpdated(address snapshotterAddress, bool allowed)
func (_ProtocolState *ProtocolStateFilterer) WatchMasterSnapshottersUpdated(opts *bind.WatchOpts, sink chan<- *ProtocolStateMasterSnapshottersUpdated) (event.Subscription, error) {

	logs, sub, err := _ProtocolState.contract.WatchLogs(opts, "masterSnapshottersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProtocolStateMasterSnapshottersUpdated)
				if err := _ProtocolState.contract.UnpackLog(event, "masterSnapshottersUpdated", log); err != nil {
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
func (_ProtocolState *ProtocolStateFilterer) ParseMasterSnapshottersUpdated(log types.Log) (*ProtocolStateMasterSnapshottersUpdated, error) {
	event := new(ProtocolStateMasterSnapshottersUpdated)
	if err := _ProtocolState.contract.UnpackLog(event, "masterSnapshottersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

//SPDX-License-Identifier: MIT
/* Copyright (c) 2023 PowerLoom, Inc. */

//Currently set to finalize epoch when 51% of total or "active" (epoch submission period is currently 10s) snapshotters have submitted the same CID. Not fully tested for all scenarios.

pragma solidity ^0.8.17;
import "hardhat/console.sol";

contract AuditRecordStoreDynamicSnapshotters {
    address payable public owner;
    uint256 public totalSnapshotterCount;
    mapping (address => bool) public snapshotters;
    mapping (string => mapping (address => bool)) public projectSnapshotters;
    mapping (string => uint256) public projectSnapshottersCount;
    mapping (string => mapping (uint256 => mapping (address => bool))) public snapshotsReceived;
    mapping (string => mapping (uint256 => mapping (string => uint256))) public snapshotsReceivedCount;
    mapping (string => mapping (uint256 => bool)) public epochStatus;
    uint256 public currentEpoch;
    mapping (uint256 => uint256) public epochTime;
    mapping (string => mapping (uint256 => uint256)) public maxSnapshotsCount;
    mapping (string => mapping (uint256 => string)) public maxSnapshotsCid;
 
    event RecordAppended(address snapshotterAddr, string snapshotCid, string payloadCommitId, uint256 tentativeBlockHeight, string projectId, uint256 indexed timestamp);
    event DelayedRecordAppended(address snapshotterAddr, string snapshotCid, string payloadCommitId, uint256 tentativeBlockHeight, string projectId, uint256 indexed timestamp);
    event EpochFinalized(uint256 tentativeBlockHeight, string projectId, string snapshotCid, uint256 indexed timestamp);
    event snapshotterRegistered(address snapshotterAddr, uint256 projectId);
    event NewEpoch(uint256 epoch);

    constructor() payable {
        owner = payable(msg.sender);
    }

    function registerPeer(address snapshotterAddr, string memory projectId) public {
        require(projectSnapshotters[projectId][snapshotterAddr] == false, "Snapshotter already exists!");
        if (!snapshotters[snapshotterAddr]){
            snapshotters[snapshotterAddr] = true;
            totalSnapshotterCount++;
        }
        projectSnapshotters[projectId][snapshotterAddr] = true;
        projectSnapshottersCount[projectId]++;
    }

    function commitRecord(string memory snapshotCid, string memory payloadCommitId, uint256 tentativeBlockHeight, string memory projectId, address snapshotterAddr) public {
        //checkConsensus(snapshotterAddr, projectId, tentativeBlockHeight);
        require(projectSnapshotters[projectId][snapshotterAddr] == true, "Snapshotter does not exist!");
        require(epochTime[tentativeBlockHeight] > 0, "Epoch does not exist!");
        require(snapshotsReceived[projectId][tentativeBlockHeight][snapshotterAddr] == false, "Snapshotter has already sent snapshot for this epoch!");
        if (epochTime[tentativeBlockHeight]+10 >= block.timestamp){
            emit RecordAppended(snapshotterAddr, snapshotCid, payloadCommitId, tentativeBlockHeight, projectId, block.timestamp);
            snapshotsReceivedCount[projectId][tentativeBlockHeight][snapshotCid]++;
            if (snapshotsReceivedCount[projectId][tentativeBlockHeight][snapshotCid] == maxSnapshotsCount[projectId][tentativeBlockHeight]){
                //equal not good
                maxSnapshotsCid[projectId][tentativeBlockHeight] = '';
                console.log('max snapshots reset due to CIDs with equal count!', projectId, tentativeBlockHeight);
            }
            if (snapshotsReceivedCount[projectId][tentativeBlockHeight][snapshotCid] > maxSnapshotsCount[projectId][tentativeBlockHeight]){
                maxSnapshotsCount[projectId][tentativeBlockHeight] = snapshotsReceivedCount[projectId][tentativeBlockHeight][snapshotCid];
                maxSnapshotsCid[projectId][tentativeBlockHeight] = snapshotCid;
                console.log('set max snapshots', projectId, tentativeBlockHeight, maxSnapshotsCid[projectId][tentativeBlockHeight]);
            }
        } else {
            console.log('Epoch submission period has passed, this submission has been marked as delayed', snapshotterAddr, projectId, tentativeBlockHeight);
            emit DelayedRecordAppended(snapshotterAddr, snapshotCid, payloadCommitId, tentativeBlockHeight, projectId, block.timestamp);
        }
        snapshotsReceived[projectId][tentativeBlockHeight][snapshotterAddr] = true;
        if (epochStatus[projectId][tentativeBlockHeight] == false){
            console.log('total check', snapshotsReceivedCount[projectId][tentativeBlockHeight][snapshotCid], projectSnapshottersCount[projectId]);
            if (snapshotsReceivedCount[projectId][tentativeBlockHeight][snapshotCid]*10 >= projectSnapshottersCount[projectId]*10/2){
                epochStatus[projectId][tentativeBlockHeight] = true;
                emit EpochFinalized(tentativeBlockHeight, projectId, snapshotCid, block.timestamp);
            } else {
                forceCompleteConsensus(projectId, tentativeBlockHeight, true);
            }
        }
    }

    function addEpoch(uint256 epoch) public {
        require(msg.sender == owner, "You need to be the owner to set epochs!");
        currentEpoch = epoch;
        epochTime[epoch] = block.timestamp;
        emit NewEpoch(epoch);
    }

    function forceCompleteConsensus(string memory projectId, uint256 tentativeBlockHeight, bool internalCheck) public {
        //require(msg.sender == owner, "You need to be the owner to force consensus!");
        if (checkDynamicConsensus(projectId, tentativeBlockHeight)){
            epochStatus[projectId][tentativeBlockHeight] = true;
            emit EpochFinalized(tentativeBlockHeight, projectId, maxSnapshotsCid[projectId][tentativeBlockHeight], block.timestamp);
        } else if (!internalCheck) {
            console.log("check failed for consensus", projectId, tentativeBlockHeight);
        }
    }

    function checkDynamicConsensus(string memory projectId, uint256 tentativeBlockHeight) public view returns (bool) {
        if (epochStatus[projectId][tentativeBlockHeight]){
            console.log('This epoch has already been finalized! Move along, move along..');
        } else {
            if (epochTime[tentativeBlockHeight]+10 < block.timestamp){
                console.log("epoch submission period passed", epochTime[tentativeBlockHeight], maxSnapshotsCid[projectId][tentativeBlockHeight], block.timestamp);
                bytes memory cidTest = bytes(maxSnapshotsCid[projectId][tentativeBlockHeight]);
                if (maxSnapshotsCount[projectId][tentativeBlockHeight] > 1 && cidTest.length > 0){
                    console.log('Consensus has reached!');
                    return true;
                } else {
                    console.log('Consensus could not be reached on time! This epoch will remain unfinalized.');
                }
            } else {
                console.log('Epoch submission period is still active!');
            }
        }
        return false;
    }
}

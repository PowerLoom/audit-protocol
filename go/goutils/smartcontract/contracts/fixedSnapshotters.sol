//SPDX-License-Identifier: MIT
/* Copyright (c) 2023 PowerLoom, Inc. */

//Currently set to finalize epoch when 51% of snapshotters have submitted the same CID - no time restriction. Not fully tested for all scenarios. 

pragma solidity ^0.8.17;

contract AuditRecordStoreFixedSnapshotters {
    address payable public owner;
    uint256 public totalSnapshotterCount;
    mapping (address => bool) public snapshotters;
    mapping (string => mapping (address => bool)) public projectSnapshotters;
    mapping (string => uint256) public projectSnapshottersCount;
    mapping (string => mapping (uint256 => mapping (address => bool))) public snapshotsReceived;
    mapping (string => mapping (uint256 => mapping (string => uint256))) public snapshotsReceivedCount;
    mapping (string => mapping (uint256 => bool)) public epochStatus;
 
    event RecordAppended(address snapshotterAddr, string snapshotCid, string payloadCommitId, uint256 tentativeBlockHeight, string projectId, uint256 indexed timestamp);
    event EpochFinalized(uint256 tentativeBlockHeight, string projectId, string snapshotCid, uint256 indexed timestamp);
    event snapshotterRegistered(address snapshotterAddr, uint256 projectId);

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
        require(snapshotsReceived[projectId][tentativeBlockHeight][snapshotterAddr] == false, "Snapshotter has already sent snapshot for this epoch!");
        emit RecordAppended(snapshotterAddr, snapshotCid, payloadCommitId, tentativeBlockHeight, projectId, block.timestamp);
        snapshotsReceived[projectId][tentativeBlockHeight][snapshotterAddr] = true;
        snapshotsReceivedCount[projectId][tentativeBlockHeight][snapshotCid]++;
        if (epochStatus[projectId][tentativeBlockHeight] == false && snapshotsReceivedCount[projectId][tentativeBlockHeight][snapshotCid]*10 >= projectSnapshottersCount[projectId]*10/2){
            epochStatus[projectId][tentativeBlockHeight] = true;
            emit EpochFinalized(tentativeBlockHeight, projectId, snapshotCid, block.timestamp);
        }
    }
}

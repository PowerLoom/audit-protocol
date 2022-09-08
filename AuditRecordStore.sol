/* Copyright (c) 2022 PowerLoom, Inc. */

pragma solidity ^0.5.17;
pragma experimental ABIEncoderV2;


contract AuditRecordStorePub {
    event RecordAppended(bytes32 apiKeyHash, string snapshotCid, string payloadCommitId, uint256 tentativeBlockHeight, string projectId, uint256 indexed timestamp);
    constructor() public {

    }
    function commitRecord(string memory snapshotCid, string memory payloadCommitId, uint256 tentativeBlockHeight, string memory projectId, bytes32 apiKeyHash) public {
        emit RecordAppended(apiKeyHash, snapshotCid, payloadCommitId, tentativeBlockHeight, projectId, now);
    }
}

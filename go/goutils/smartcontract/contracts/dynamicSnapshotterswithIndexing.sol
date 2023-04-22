//SPDX-License-Identifier: MIT
/* Copyright (c) 2023 PowerLoom, Inc. */

//Currently set to finalize epoch when 51% of total or "active" (epoch submission period is currently 10s) snapshotters have submitted the same CID. Not fully tested for all scenarios.

pragma solidity 0.8.17;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

// import hardhat console
contract StringSet {
    mapping(string => bool) private mySet;
    string[] private keys;

    function add(string memory value) public {
        require(!mySet[value], "Value already exists in set");
        mySet[value] = true;
        keys.push(value);
    }

    function remove(string memory value) public {
        require(mySet[value], "Value does not exist in set");
        mySet[value] = false;
        for (uint256 i = 0; i < keys.length; i++) {
            if (keccak256(bytes(keys[i])) == keccak256(bytes(value))) {
                keys[i] = keys[keys.length - 1];
                keys.pop();
                break;
            }
        }
    }

    function contains(string memory value) public view returns(bool) {
        return mySet[value];
    }

    function getAll() public view returns(string[] memory) {
        return keys;
    }
}

contract AuditRecordStoreDynamicSnapshottersWithIndexing is Ownable, EIP712 {
    using ECDSA
    for bytes32;

    using EnumerableSet
    for EnumerableSet.AddressSet;
    struct Epoch {
        uint256 begin;
        uint256 end;
    }

    struct ConsensusStatus {
        bool finalized;
        uint256 timestamp;
    }
    struct IndexAgainstHeadDAGBlock {
        uint256 tailDAGBlockHeight;
        uint256 tailDAGBlockEpochSourceChainHeight;
    }

    struct ReleaseInfo {
        uint256 timestamp;
        uint256 blocknumber;
    }

    struct Request {
        uint256 deadline;
    }

    bytes32 private constant REQUEST_TYPEHASH = keccak256("Request(uint256 deadline)");


    Epoch public currentEpoch;
    StringSet private projectSet;
    EnumerableSet.AddressSet private snapshotterSet;
    EnumerableSet.AddressSet private operatorSet;

    // NOTE: necessary to maintain epoch size per project ID or assume uniform size across projects?
    // -> Would be better to start with one epoch size for now
    uint8 private constant EPOCH_SIZE = 10;
    uint256 public totalSnapshotterCount;
    uint256 public snapshotSubmissionWindow = 1; // number of blocks to wait before finalizing epoch
    uint256 public indexSubmissionWindow = 1; // number of blocks to wait before finalizing index
    uint256 public aggregateSubmissionWindow = 1; // number of blocks to wait before finalizing aggregate

    uint256 public minSubmissionsForConsensus = 2;

    mapping(uint256 => ReleaseInfo) public epochReleaseTime;
    mapping(uint256 => ReleaseInfo) public indexStartTime;
    mapping(uint256 => ReleaseInfo) public aggregateStartTime;

    mapping(string => mapping(uint256 => uint256)) public maxSnapshotsCount;
    mapping(string => mapping(uint256 => string)) public maxSnapshotsCid;

    mapping(string => mapping(uint256 => uint256)) public maxAggregatesCount;
    mapping(string => mapping(uint256 => string)) public maxAggregatesCid;

    // maxIndexCount
    mapping(string => mapping(bytes32 => mapping(uint256 => uint256))) public maxIndexCount;
    // maxIndexData
    mapping(string => mapping(bytes32 => mapping(uint256 => IndexAgainstHeadDAGBlock))) public maxIndexData;


    mapping(address => bool) public snapshotters;
    mapping(string => mapping(uint256 => mapping(address => bool))) public snapshotsReceived;
    mapping(string => mapping(uint256 => mapping(address => bool))) public aggregateReceived;
    mapping(string => mapping(uint256 => mapping(string => uint256))) public snapshotsReceivedCount;
    mapping(string => mapping(uint256 => mapping(string => uint256))) public aggregateReceivedCount;

    // project ID -> DAG block height to epoch finalization status. epoch_end at dagHeight = (dag_block_height - 1) * EPOCH_SIZE + projectFirstEpochEndHeight[projectId]
    // projectid -> dagHeight -> ConsensusStatus
    mapping(string => mapping(uint256 => ConsensusStatus)) public snapshotStatus;
    mapping(string => mapping(uint256 => ConsensusStatus)) public aggregateStatus;

    mapping(bytes32 => string) private TIME_SERIES_INDEX_IDENTIFIERS;
    mapping(string => uint256) public projectFirstEpochEndHeight;
    // raw snapshot submissions
    // projectId -> dagBlockHeight -> dagCid
    mapping(string => mapping(uint256 => string)) public finalizedDagCids;
    // updated index submissions
    // project ID -> time series identifier -> DAG block height to index -> indexer address to submission status
    mapping(string => mapping(bytes32 => mapping(uint256 => mapping(address => bool)))) public indexReceived;
    // NOTE: drawback of counting submissions against a hash of dagTailHeight+sourceChainEpochHeight: in case of no consensus being reached, everyone's submissions would not be known explicitly
    // project ID -> time series identifier -> DAG block height to index -> indexer address to submission count
    mapping(string => mapping(bytes32 => mapping(uint256 => mapping(bytes32 => uint256)))) public indexReceivedCount;

    // project ID -> time series identifier -> DAG block height to index -> indexer address to submission count
    mapping(string => mapping(bytes32 => mapping(uint256 => IndexAgainstHeadDAGBlock))) public finalizedIndexes;

    // project ID -> time series identifier -> DAG block height to index finalization status
    mapping(string => mapping(bytes32 => mapping(uint256 => ConsensusStatus))) public indexStatus;

    event SnapshotterAllowed(address snapshotterAddress);
    event SnapshotterRegistered(address snapshotterAddr, string projectId);
    event StateBuilderAllowed(address stateBuilderAddress);
    // to be set in constructor as initial address
    event AdminModuleUpdated(address adminModuleAddress);

    // TODO: Update EpochReleased event to generate unique Epoch ID
    // Context: Ditch local DAGBlockHeight and everything related, this data will come from Epoch but will be named the same
    // TODO: !IMPORTANT - Build commit and reveal mechansim in contract (update events and systems according to this upon discussion with team)

    //Epoch to be released from Protocol Admin Module
    event EpochReleased(uint256 begin, uint256 end, uint256 indexed timestamp);
    // snapshotters listen to this even, generate snapshots and submit to contract

    event SnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 epochEnd, string projectId, uint256 indexed timestamp);

    event DelayedSnapshotSubmitted(address snapshotterAddr, string snapshotCid, uint256 epochEnd, string projectId, uint256 indexed timestamp);

    event SnapshotFinalized(uint256 epochEnd, string projectId, string snapshotCid, uint256 indexed timestamp);

    // State builders (initally only run by Powerloom) will listen to EpochFinalized events
    // it will read finalized snapshot and build dag chain by referencing current and previous snapshot then submit to contract
    event SnapshotDagCidSubmitted(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp);
    // if consensus achieved emit DagCidFinalized
    event SnapshotDagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp);

    // Index
    // Snapshotters will listen to EpochFinalized Event and start indexing for that (epoch, project)
    // on submission release
    event IndexSubmitted(address snapshotterAddr, uint256 indexTailDAGBlockHeight, uint256 DAGBlockHeight, string projectId, bytes32 indexIdentifierHash, uint256 indexed timestamp);
    // in case index submission window is closed
    event DelayedIndexSubmitted(address snapshotterAddr, uint256 indexTailDAGBlockHeight, uint256 DAGBlockHeight, string projectId, bytes32 indexIdentifierHash, uint256 indexed timestamp);
    // on Index Finalization (consensus) release
    event IndexFinalized(string projectId, uint256 DAGBlockHeight, uint256 indexTailDAGBlockHeight, uint256 tailBlockEpochSourceChainHeight, bytes32 indexIdentifierHash, uint256 indexed timestamp);
    // Aggregation
    // TODO: Recheck events below and update if needed
    event AggregateSubmitted(address snapshotterAddr, string aggregateCid, uint256 epochEnd, string projectId, uint256 indexed timestamp);
    // renamed to DelayerAggregateSubmitted (emitted if submitted after submission window)
    event DelayedAggregateSubmitted(address snapshotterAddr, string aggregateCid, uint256 epochEnd, string projectId, uint256 indexed timestamp);
    // on Aggreate Finalization
    event AggregateFinalized(uint256 epochEnd, string projectId, string aggregateCid, uint256 indexed timestamp);

    // State builders (initally only run by Powerloom) will listen to AggregateFinalized events
    // it will read finalized snapshot and build dag chain by referencing current and previous snapshot then submit to contract
    event AggregateDagCidSubmitted(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp);
    // if consensus achieved emit DagCidFinalized
    event AggregateDagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp);

    event DagCidFinalized(string projectId, uint256 indexed DAGBlockHeight, string dagCid, uint256 indexed timestamp);
    modifier onlyOperator {
        require(operatorSet.contains(msg.sender), "Only operator can call this function.");
        _;
    }

    modifier onlyOwnerOrOperator {
        require(owner() == msg.sender || operatorSet.contains(msg.sender), "Only owner or operator can call this function!");
        _;
    }

    constructor()
    EIP712("PowerloomProtocolContract", "0.1") {
        projectSet = new StringSet();
        // TODO: Time series index identifiers will be made dynamic later
        TIME_SERIES_INDEX_IDENTIFIERS[keccak256("24h")] = "24h";
        TIME_SERIES_INDEX_IDENTIFIERS[keccak256("7d")] = "7d";
    }

    function verify(
        Request calldata request,
        bytes calldata signature,
        address signer
    ) public view returns(bool) {
        bytes32 requestHash = hashRequest(request);
        require(signer == recoverAddress(requestHash, signature), "Invalid signature");
        require(block.number < request.deadline, "Signature Expired!");
        return true;
    }

    function hashRequest(Request calldata request) private view returns(bytes32) {

        bytes32 requestHash = _hashTypedDataV4(keccak256(
            abi.encode(
                REQUEST_TYPEHASH,
                request.deadline
            )
        ));
        return requestHash;
    }

    function recoverAddress(bytes32 messageHash, bytes calldata signature) public pure returns(address) {
        return messageHash.recover(signature);
    }

    function allowSnapshotter(address snapshotterAddr) public onlyOwnerOrOperator {
        require(!snapshotters[snapshotterAddr], "Snapshotter already allowed!");
        snapshotters[snapshotterAddr] = true;
        if (!snapshotterSet.contains(snapshotterAddr)) {
            snapshotterSet.add(snapshotterAddr);
        }
        emit SnapshotterAllowed(snapshotterAddr);
    }

    function getProjects() public view returns(string[] memory) {
        return projectSet.getAll();
    }

    function addOperator(address operatorAddr) public onlyOwner {
        require(operatorSet.contains(operatorAddr) == false, "Operator already exists!");
        operatorSet.add(operatorAddr);
    }

    function updateMinSnapshottersForConsensus(uint256 _minSubmissionsForConsensus) public onlyOwner {
        minSubmissionsForConsensus = _minSubmissionsForConsensus;
    }

    function removeOperator(address operatorAddr) public onlyOwner {
        require(operatorSet.contains(operatorAddr) == true, "Operator does not exist!");
        operatorSet.remove(operatorAddr);
    }

    function getOperators() public view returns(address[] memory) {
        return operatorSet.values();
    }

    function getAllSnapshotters() public view returns(address[] memory) {
        return snapshotterSet.values();
    }

    function getTotalSnapshotterCount() public view returns(uint256) {
        return snapshotterSet.length();
    }

    function updateSnapshotSubmissionWindow(uint256 newsnapshotSubmissionWindow) public onlyOwner {
        snapshotSubmissionWindow = newsnapshotSubmissionWindow;
    }

    function updateIndexSubmissionWindow(uint256 newIndexSubmissionWindow) public onlyOwner {
        indexSubmissionWindow = newIndexSubmissionWindow;
    }

    function updateAggregateSubmissionWindow(uint256 newAggregateSubmissionWindow) public onlyOwner {
        aggregateSubmissionWindow = newAggregateSubmissionWindow;
    }

    function releaseEpoch(uint256 begin, uint256 end) public onlyOwnerOrOperator{

        require(end > begin, "Epoch end must be greater than begin!");
        currentEpoch = Epoch(begin, end);
        epochReleaseTime[end] = ReleaseInfo(block.timestamp, block.number);
        emit EpochReleased(begin, end, block.timestamp);
    }


    function submitSnapshot(string memory snapshotCid, uint256 epochEnd, string memory projectId, Request calldata request, bytes calldata signature) public onlyOwner {
        address snapshotterAddr = recoverAddress(hashRequest(request), signature);
        require(verify(request, signature, snapshotterAddr), "Can't verify signature");
        require(snapshotters[snapshotterAddr] == true, "Snapshotter does not exist!");

        require(epochReleaseTime[epochEnd].timestamp > 0, "Epoch does not exist!");
        require(snapshotsReceived[projectId][epochEnd][snapshotterAddr] == false, "Snapshotter has already sent snapshot for this epoch!");
        
        // Update project list
        if (!projectSet.contains(projectId)) {
            projectSet.add(projectId);
        }

        if (epochReleaseTime[epochEnd].blocknumber + snapshotSubmissionWindow >= block.number) {

            if (projectFirstEpochEndHeight[projectId] == 0) {
                projectFirstEpochEndHeight[projectId] = epochEnd;
            }

            emit SnapshotSubmitted(snapshotterAddr, snapshotCid, epochEnd, projectId, block.timestamp);
            snapshotsReceivedCount[projectId][epochEnd][snapshotCid]++;
            if (snapshotsReceivedCount[projectId][epochEnd][snapshotCid] == maxSnapshotsCount[projectId][epochEnd]) {
                //equal not good
                maxSnapshotsCid[projectId][epochEnd] = '';
            }
            if (snapshotsReceivedCount[projectId][epochEnd][snapshotCid] > maxSnapshotsCount[projectId][epochEnd]) {
                maxSnapshotsCount[projectId][epochEnd] = snapshotsReceivedCount[projectId][epochEnd][snapshotCid];
                maxSnapshotsCid[projectId][epochEnd] = snapshotCid;

            }
        } else {
            emit DelayedSnapshotSubmitted(snapshotterAddr, snapshotCid, epochEnd, projectId, block.timestamp);
        }
        snapshotsReceived[projectId][epochEnd][snapshotterAddr] = true;
        if (snapshotStatus[projectId][epochEnd].finalized == false) {
            if (
                snapshotsReceivedCount[projectId][epochEnd][snapshotCid] * 10 >= getTotalSnapshotterCount() * 10 / 2 &&
                maxSnapshotsCount[projectId][epochEnd] >= minSubmissionsForConsensus) {
                snapshotStatus[projectId][epochEnd].finalized = true;
                snapshotStatus[projectId][epochEnd].timestamp = block.timestamp;

                emit SnapshotFinalized(epochEnd, projectId, snapshotCid, block.timestamp);

                // Starting indexing and aggregation windows
                indexStartTime[epochEnd] = ReleaseInfo(block.timestamp, block.number);
                aggregateStartTime[epochEnd] = ReleaseInfo(block.timestamp, block.number);

            } else {
                forceCompleteConsensusSnapshot(projectId, epochEnd);
            }
        }
    }


    function submitIndex(
        string memory projectId,
        uint256 DAGBlockHeight, // DAG block height, head of DAG against which tail of time series index is committed
        uint256 indexTailDAGBlockHeight,
        uint256 tailBlockEpochSourceChainHeight,
        bytes32 indexIdentifierHash, // hash of time series identifiers: "24h", "7d"
        Request calldata request,
        bytes calldata signature
    ) public onlyOwner {
        address snapshotterAddr = recoverAddress(hashRequest(request), signature);
        require(verify(request, signature, snapshotterAddr), "Can't verify signature");

        require(indexStartTime[DAGBlockHeight].timestamp > 0, "Indexing not open for this epoch!");
        require(keccak256(abi.encode(TIME_SERIES_INDEX_IDENTIFIERS[indexIdentifierHash])) != keccak256(abi.encode("")), "Index identifier does not exist!");

        require(snapshotters[snapshotterAddr] == true, "Snapshotter does not exist!");
        require(indexReceived[projectId][indexIdentifierHash][DAGBlockHeight][snapshotterAddr] == false, "Snapshotter has already sent index for this epoch!");

        bytes32 dagTailEpochHeightHash = keccak256(abi.encode(indexTailDAGBlockHeight, tailBlockEpochSourceChainHeight));

        if (indexStartTime[DAGBlockHeight].blocknumber + indexSubmissionWindow >= block.number) {
            emit IndexSubmitted(snapshotterAddr, indexTailDAGBlockHeight, DAGBlockHeight, projectId, indexIdentifierHash, block.timestamp);


            indexReceivedCount[projectId][indexIdentifierHash][DAGBlockHeight][dagTailEpochHeightHash]++;

            if (indexReceivedCount[projectId][indexIdentifierHash][DAGBlockHeight][dagTailEpochHeightHash] == maxIndexCount[projectId][indexIdentifierHash][DAGBlockHeight]) {
                //equal not good
                maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockHeight = 0;
                maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockEpochSourceChainHeight = 0;
            }

            if (indexReceivedCount[projectId][indexIdentifierHash][DAGBlockHeight][dagTailEpochHeightHash] > maxIndexCount[projectId][indexIdentifierHash][DAGBlockHeight]) {
                maxIndexCount[projectId][indexIdentifierHash][DAGBlockHeight] = indexReceivedCount[projectId][indexIdentifierHash][DAGBlockHeight][dagTailEpochHeightHash];
                maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockHeight = indexTailDAGBlockHeight;
                maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockEpochSourceChainHeight = tailBlockEpochSourceChainHeight;
            }

        } else {
            emit DelayedIndexSubmitted(snapshotterAddr, indexTailDAGBlockHeight, DAGBlockHeight, projectId, indexIdentifierHash, block.timestamp);

        }

        indexReceived[projectId][indexIdentifierHash][DAGBlockHeight][snapshotterAddr] = true;

        if (!indexStatus[projectId][indexIdentifierHash][DAGBlockHeight].finalized) {

            if (
                indexReceivedCount[projectId][indexIdentifierHash][DAGBlockHeight][dagTailEpochHeightHash] * 10 >= getTotalSnapshotterCount() * 10 / 2 &&
                maxIndexCount[projectId][indexIdentifierHash][DAGBlockHeight] >= minSubmissionsForConsensus) {
                indexStatus[projectId][indexIdentifierHash][DAGBlockHeight].finalized = true;
                indexStatus[projectId][indexIdentifierHash][DAGBlockHeight].timestamp = block.timestamp;
                finalizedIndexes[projectId][indexIdentifierHash][DAGBlockHeight] = IndexAgainstHeadDAGBlock(indexTailDAGBlockHeight, tailBlockEpochSourceChainHeight);
                emit IndexFinalized(projectId, DAGBlockHeight, indexTailDAGBlockHeight, tailBlockEpochSourceChainHeight, indexIdentifierHash, block.timestamp);
            } else {
                forceCompleteConsensusIndex(projectId, DAGBlockHeight, indexIdentifierHash);
            }
        }
    }

    function submitAggregate(
        string memory snapshotCid,
        uint256 DAGBlockHeight,
        string memory projectId,
        Request calldata request,
        bytes calldata signature
    ) public onlyOwner {
        address snapshotterAddr = recoverAddress(hashRequest(request), signature);
        require(verify(request, signature, snapshotterAddr), "Can't verify signature");
        // check if snapshotterAddress is in snapshotters
        require(snapshotters[snapshotterAddr] == true, "Snapshotter does not exist!");
        require(epochReleaseTime[DAGBlockHeight].timestamp > 0, "Epoch does not exist!");
        require(aggregateStartTime[DAGBlockHeight].timestamp > 0, "Aggregation not open for this epoch!");
        require(aggregateReceived[projectId][DAGBlockHeight][snapshotterAddr] == false, "Snapshotter has already sent aggregate for this epoch!");
        
        if (!projectSet.contains(projectId)) {
            projectSet.add(projectId);
        }
        
        if (aggregateStartTime[DAGBlockHeight].blocknumber + aggregateSubmissionWindow >= block.number) {
            emit AggregateSubmitted(snapshotterAddr, snapshotCid, DAGBlockHeight, projectId, block.timestamp);
            aggregateReceivedCount[projectId][DAGBlockHeight][snapshotCid]++;

            if (aggregateReceivedCount[projectId][DAGBlockHeight][snapshotCid] == maxAggregatesCount[projectId][DAGBlockHeight]) {
                //equal not good
                maxAggregatesCid[projectId][DAGBlockHeight] = '';
            }
            if (aggregateReceivedCount[projectId][DAGBlockHeight][snapshotCid] > maxAggregatesCount[projectId][DAGBlockHeight]) {
                maxAggregatesCount[projectId][DAGBlockHeight] = aggregateReceivedCount[projectId][DAGBlockHeight][snapshotCid];
                maxAggregatesCid[projectId][DAGBlockHeight] = snapshotCid;
            }
        } else {
            emit DelayedAggregateSubmitted(snapshotterAddr, snapshotCid, DAGBlockHeight, projectId, block.timestamp);
        }

        aggregateReceived[projectId][DAGBlockHeight][snapshotterAddr] = true;
        if (aggregateStatus[projectId][DAGBlockHeight].finalized == false) {
            if (
                aggregateReceivedCount[projectId][DAGBlockHeight][snapshotCid] * 10 >= getTotalSnapshotterCount() * 10 / 2 &&
                maxAggregatesCount[projectId][DAGBlockHeight] >= minSubmissionsForConsensus) {
                aggregateStatus[projectId][DAGBlockHeight].finalized = true;
                aggregateStatus[projectId][DAGBlockHeight].timestamp = block.timestamp;

                emit AggregateFinalized(DAGBlockHeight, projectId, snapshotCid, block.timestamp);
            } else {
                forceCompleteConsensusAggregate(projectId, DAGBlockHeight);
            }
        }
    }



    function forceCompleteConsensusSnapshot(string memory projectId, uint256 epochEnd) public {
  
        if (checkDynamicConsensusSnapshot(projectId, epochEnd)) {
            snapshotStatus[projectId][epochEnd].finalized = true;
            snapshotStatus[projectId][epochEnd].timestamp = block.timestamp;

            // update window of DAG block heights against which indexing is open to being accepted
            indexStartTime[epochEnd] = ReleaseInfo(block.timestamp, block.number);
            aggregateStartTime[epochEnd] = ReleaseInfo(block.timestamp, block.number);

            emit SnapshotFinalized(epochEnd, projectId, maxSnapshotsCid[projectId][epochEnd], block.timestamp);
        }
    }

    function forceCompleteConsensusAggregate(string memory projectId, uint256 DAGBlockHeight) public {

        if (checkDynamicConsensusAggregate(projectId, DAGBlockHeight) && aggregateStatus[projectId][DAGBlockHeight].finalized == false) {
            aggregateStatus[projectId][DAGBlockHeight].finalized = true;
            aggregateStatus[projectId][DAGBlockHeight].timestamp = block.timestamp;

            emit AggregateFinalized(DAGBlockHeight, projectId, maxAggregatesCid[projectId][DAGBlockHeight], block.timestamp);
        }
    }

    function forceCompleteConsensusIndex(string memory projectId, uint256 DAGBlockHeight, bytes32 indexIdentifierHash) public {

        if (checkDynamicConsensusIndex(projectId, DAGBlockHeight, indexIdentifierHash) && indexStatus[projectId][indexIdentifierHash][DAGBlockHeight].finalized == false) {
            indexStatus[projectId][indexIdentifierHash][DAGBlockHeight].finalized = true;
            indexStatus[projectId][indexIdentifierHash][DAGBlockHeight].timestamp = block.timestamp;
            finalizedIndexes[projectId][indexIdentifierHash][DAGBlockHeight] = IndexAgainstHeadDAGBlock(maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockHeight, maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockEpochSourceChainHeight);
            
            emit IndexFinalized(projectId, DAGBlockHeight, maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockHeight, maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockEpochSourceChainHeight, indexIdentifierHash, block.timestamp);
        }
    }

    function checkDynamicConsensusSnapshot(string memory projectId, uint256 epochEnd) public view returns(bool success) {
        if (!snapshotStatus[projectId][epochEnd].finalized) {
            {
                if (epochReleaseTime[epochEnd].blocknumber + snapshotSubmissionWindow < block.number) {

                    bytes memory cidTest = bytes(maxSnapshotsCid[projectId][epochEnd]);
                    if (maxSnapshotsCount[projectId][epochEnd] >= minSubmissionsForConsensus && cidTest.length > 0) {
                        return true;
                    }
                }
            }
            return false;
        }
    }


    function checkDynamicConsensusIndex(string memory projectId, uint256 DAGBlockHeight, bytes32 indexIdentifierHash) public view returns(bool success) {
        if (!indexStatus[projectId][indexIdentifierHash][DAGBlockHeight].finalized) {
            {
                if (indexStartTime[DAGBlockHeight].blocknumber + indexSubmissionWindow < block.number) {

                    if (maxIndexCount[projectId][indexIdentifierHash][DAGBlockHeight] >= minSubmissionsForConsensus && maxIndexData[projectId][indexIdentifierHash][DAGBlockHeight].tailDAGBlockHeight != 0) {
                        return true;
                    }
                }
            }
            return false;
        }
    }


    function checkDynamicConsensusAggregate(string memory projectId, uint256 DAGBlockHeight) public view returns(bool success) {
        if (!aggregateStatus[projectId][DAGBlockHeight].finalized) {
            {
                if (aggregateStartTime[DAGBlockHeight].blocknumber + aggregateSubmissionWindow < block.number) {

                    bytes memory cidTest = bytes(maxAggregatesCid[projectId][DAGBlockHeight]);
                    if (maxAggregatesCount[projectId][DAGBlockHeight] >= minSubmissionsForConsensus && cidTest.length > 0) {
                        return true;
                    }
                }
            }
            return false;
        }
    }

    function commitFinalizedDAGcid(string memory projectId, uint256 dagBlockHeight, string memory dagCid, Request calldata request, bytes calldata signature) public {
        address signerAddr = recoverAddress(hashRequest(request), signature);
        require(verify(request, signature, signerAddr), "Can't verify signature");
        require(owner() == signerAddr || operatorSet.contains(signerAddr), "Only owner or operator can call this function!");
        
        require(snapshotStatus[projectId][dagBlockHeight].finalized == true);
        finalizedDagCids[projectId][dagBlockHeight] = dagCid;
        emit DagCidFinalized(projectId, dagBlockHeight, dagCid, block.timestamp);
    }

}
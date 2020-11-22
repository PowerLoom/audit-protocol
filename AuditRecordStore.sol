pragma solidity ^0.5.17;
pragma experimental ABIEncoderV2;

contract AuditRecordStore {
    struct PayloadRecord {
        string ipfsCid;
        uint256 timestamp;
    }

    event RecordAppended(bytes32 apiKeyHash, string ipfsCid, uint256 indexed timestamp);

    mapping(bytes32 => PayloadRecord[]) private apiKeyHashToRecords;


    constructor() public {

    }

    function commitRecord(string memory ipfsCid, bytes32 apiKeyHash) public {
        PayloadRecord memory a = PayloadRecord(ipfsCid, now);
        apiKeyHashToRecords[apiKeyHash].push(a);
        emit RecordAppended(apiKeyHash, ipfsCid, now);
    }

    /*
    function getTokenRecordLogs(bytes32 tokenHash) public view returns (PayloadRecord[] memory recordLogs) {
        return tokenHashesToRecords[tokenHash];
    }
    */
}
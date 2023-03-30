# DAG Verifier Service Analysis Report

- main()
    - Initializing DagVerifier Struct
    - creating http listener for `reportIssue` endpoint
    - Initializing pruning verifier
    - Run DagVerifier
    - Run PruningVerifier

- `reportIssue` handler func
    - Receives issues over http. ReqBody is expected to
      be [IssueReport](https://github.com/swagftw/audit-protocol/blob/main/go/goutils/datamodel/data_model.go#L87)
    - Store report in redis
    - Report issue to consensus system
    - Remove issueReports older than 7 days

## Dag Verifier

- Initialize DagVerifier Struct
    - init slack workflow client
    - populate projects summary projects separately as well
    - FetchLastVerificationStatusFromRedis() fetches last verified dag height for every project from redis
    - FetchLastProjectIndexedStatusFromRedis() fetches hash of projectID to start source chain height and current source
      chain height


- dagVerifier.Run()
    - Again called FetchLastProjectIndexedStatusFromRedis() which was already done in Initialize DagVerifier Struct.
      Okay fine can be removed from initialize struct method
    - Again called FetchLastVerificationStatusFromRedis() same as above method
    - VerifyAllProjects() calls VerifyDagChain() for every project in separate go routines
    - SummarizeDAGIssuesAndNotify() summarizes all the DAG chain issues and sends notification (report) to slack
    - UpdateLastStatusToRedis() updates last verified dag height for every project in redis


- VerifyDagChain(projectID)
    - startScore = lastVerifiedDagBlockHeight for the projectID
    - get dag chain blocks from redis (dag block cids with height) starting from startScore to latest dag block height
    - get payload cids from redis for the project
        - if payload cids height has duplicate entries
        - get dag dag blocks for cids at duplicate heights from ipfs
        - check if cids from ipfs blocks are same
            - if same remove duplicate entry from local variable dagChain and restart loop
            - else return error for duplicate entries
        - check if payload cid is null, if yes remove that entry from local variable dagChain and restart loop
    - get payload data from local drive for payload cids
    - verify dag blocks' payload data
        - checks for duplicate chain range and gaps


- SummarizeDAGIssuesAndNotify()
    - check if dag chain is stuck for any project (Should not belong in this method)
    - create report for all the issues notify on slack
    - if issues are resolved notify on slack


- UpdateLastStatusToRedis()
    - update last verified dag height for every project in redis (projects:dagVerificationStatus)
    - update last indexed state for each project in redis (projects:IndexStatus)

## Pruning Verifier

Verifies pruning and archival of dag segments. This runs in a separate go routine in parallel with dag verifier,
if `pruning_verification` is set to `true` in config/settings.

- Init() initializes PruningVerifier struct
    - Initialize redis client
    - Initialize web3 storage client

- Run() runs pruning verifier
    - Get all projects stored in redis
    - Fetch last pruning verification status from redis
    - VerifyPruningAndArchival() verify pruning and archival

- VerifyPruningAndArchival()
    - check storage status for every dag segment
    - if dag segment has `COLD` status check stat of dag segment is uploaded to web3 storage
        - if yes verify dag segment uploaded to web3 storage (check for gaps and duplicate chain ranges)
        - else update pruning report with failed status
    - update verification status in redis stored in hash with key `projects:pruningVerificationStatus`
    - notify on slack if fails are detected with archival and pruning

## How to better verify DAG Chain

- Check if dag chain is stuck for any project
- Check for duplicate dag segments (same dag blocks range but different CID)
- Check for gaps in dag chain
- For consecutive dag blocks verify payload data
    - check if source chain is continuous (no gaps). nth block's source chain end height = (n-1)th block's source chain
      height - 1. Which also means checks for duplication of Epoch in payload data.

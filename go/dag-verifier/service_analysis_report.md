# DAG Verifier Service Analysis Report

- main()
    - Initializing DagVerifier Struct
    - creating http listener for `reportIssue` endpoint
    - Initializing pruning verifier
    - Run DagVerifier
    - Run PruningVerifier

- `reportIssue` handler func
  - Receives issues over http. ReqBody is expected to be [IssueReport](https://github.com/swagftw/audit-protocol/blob/main/go/goutils/datamodel/data_model.go#L87)
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


## Pruning Verifier (WIP)

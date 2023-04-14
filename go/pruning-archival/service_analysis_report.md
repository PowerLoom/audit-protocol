# Service Code Analysis Report

## High level findings

- Code has lots of global variables. This makes it hard to test and debug.
- Lack of comments for some functions causing misinterpretation function's functionality.
- Complete TODOs in the code.
- Optimise the code for better performance where ever mentioned in the code.

## Flow of the code:

1. [main()](https://github.com/PowerLoom/audit-protocol/blob/main/go/pruning-archival/main.go#L79) function is the entry point of the code. 
   - Initialisations such as logger, settings and other variables (global)
   - Lots of global variables are exposed and as said earlier, it makes it hard to test and debug.
   - Call to `Run()` function which is the entry point for service functionality.

2. [Run()](https://github.com/PowerLoom/audit-protocol/blob/main/go/pruning-archival/main.go#L129) function runs the service functionality
   - Runs infinite loop (each run is a cycle) which gets projects from the cache, verifies then pruning status and prune the project DAG chains
   - It gets last pruned status of all the projects from cache along with the last pruned block number (GetLastPrunedStatusFromRedis())
   - Verify the current status of the project DAG chains and prune (VerifyAndPruneDAGChains())

3. [ProcessProject()](https://github.com/PowerLoom/audit-protocol/blob/main/go/pruning-archival/main.go#L199)
   - This function is called for each project in the cycle
   - logic:
     - Get all dagSegments for the project (FetchProjectMetaData())
     - FindPruningHeight() - Find the height at which the project needs to be pruned (need more clarity on this function)
     - For each dagSegment, if the dagSegment end height is less than the pruning height check if it is in pruning pending state
     - If the dagSegment is in pending archival/pruning state, then archive the dagSegment (ArchiveDAG())
     - After archiving segment unpin CIDs from IPFS (UnPinCidsFromIPFS())
     - Backup redis cache payloadCIDs and CIDs to local file
     - Remove redis cache CIDs (PruneProjectInRedis())
     - Delete DAG CIDs from local file system (DeleteContentFromLocalCache())
     - Update pruned status in redis cache (UpdatePrunedStatusToRedis())

4. [ArchiveDAG()](https://github.com/PowerLoom/audit-protocol/blob/main/go/pruning-archival/main.go#L313)
   - Export dag from IPFS (ExportDAGFromIPFS())
   - Store CAR file in local file system
   - Upload file to web3 storage (UploadFileToWeb3Storage())
   - Remove file from local file system

## Improvements to be done (can be done):

- Complete TODOs/FIXMEs/BUGs in the code
- Add more comments to the functions to make it more readable
- Proper spacing and new lines in the code for better readability
- Using proper function names will make the code more readable (multiple functions are there like this)
- Remove need for global variables
- Need for less stored state in the service (need to think through this)
- Pass less/precise arguments to functions (than passing complete struct which contains all the variables but only single is used inside function)
- Linting the code
- There are multiple retry logic in the code, can be used centralized retry mechanism
- defer is being called inside a for loop at multiple places (must improve can cause resource leak)
- Upload file to web3 can be done in batches/concurrently (currently it is done one by one)
- IPFS Unpin can be done in batches 
- Currently if project has multiple segments to pruned (PENDING STATE).. it will processProject() only pruning on segment at a time, and waiting for another cycle 
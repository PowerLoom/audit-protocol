package redisutils

const (
	REDIS_KEY_STORED_PROJECTS                   string = "storedProjectIds"
	REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS string = "projectID:%s:unfinalizedSnapshots"
	REDIS_KEY_SNAPSHOTTER_STATUS_REPORT         string = "projectID:%s:snapshotterStatusReport"
	REDIS_KEY_TOTAL_MISSED_SNAPSHOT_COUNT       string = "projectID:%s:totalMissedSnapshotCount"
	REDIS_KEY_TOTAL_SUCCESSFUL_SNAPSHOT_COUNT   string = "projectID:%s:totalSuccessfulSnapshotCount"
	REDIS_KEY_TOTAL_INCORRECT_SNAPSHOT_COUNT    string = "projectID:%s:totalIncorrectSnapshotCount"
)

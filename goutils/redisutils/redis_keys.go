package redisutils

const REDIS_KEY_STORED_PROJECTS string = "storedProjectIds"
const REDIS_KEY_PROJECT_PAYLOAD_CIDS string = "projectID:%s:payloadCids"
const REDIS_KEY_PROJECT_CIDS string = "projectID:%s:Cids"
const REDIS_KEY_PROJECT_FINALIZED_HEIGHT string = "projectID:%s:blockHeight"
const REDIS_KEY_PRUNING_CYCLE_DETAILS string = "pruningRunStatus"

const REDIS_KEY_PRUNING_CYCLE_PROJECT_DETAILS string = "pruningProjectDetails:%s"

const REDIS_KEY_PRUNING_STATUS string = "projects:pruningStatus"
const REDIS_KEY_PROJECT_METADATA string = "projectID:%s:dagSegments"

const REDIS_KEY_PROJECT_TAIL_INDEX string = "projectID:%s:slidingCache:%s:tail"
const REDIS_KEY_PRUNING_VERIFICATION_STATUS string = "projects:pruningVerificationStatus"

const REDIS_KEY_PROJECT_EPOCH_SIZE string = "projectID:%s:epochSize"

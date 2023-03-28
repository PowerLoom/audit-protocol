package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/redisutils"
)

type RedisCache struct {
	redisClient *redis.Client
}

var _ DbCache = (*RedisCache)(nil)

func NewRedisCache() *RedisCache {
	client, err := gi.Invoke[*redis.Client]()
	if err != nil {
		log.Fatal("Failed to invoke redis client", err)
	}

	cache := &RedisCache{redisClient: client}

	err = gi.Inject(cache)
	if err != nil {
		log.Fatal("Failed to inject redis cache", err)
	}

	return cache
}

func (r *RedisCache) GetLastProjectIndexedState(ctx context.Context) (map[string]*datamodel.ProjectIndexedState, error) {
	key := redisutils.REDIS_KEY_PROJECTS_INDEX_STATUS
	indexedStateMap := make(map[string]*datamodel.ProjectIndexedState)

	val, err := r.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Errorf("error getting last project indexed state, key %s not found", key)

			return nil, ErrNotFound
		}

		return nil, ErrGettingProjects
	}

	for projectID, state := range val {
		indexedStateMap[projectID] = new(datamodel.ProjectIndexedState)

		err = json.Unmarshal([]byte(state), indexedStateMap[projectID])
		if err != nil {
			log.Errorf("error unmarshalling project indexed state for project %s", projectID)

			return nil, err
		}
	}

	return indexedStateMap, nil
}

func (r *RedisCache) GetPayloadCidAtDAGHeight(ctx context.Context, projectID string, dagHeight int) (string, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID)
	payloadCid := ""
	height := strconv.Itoa(int(dagHeight))

	log.Debug("Geting PayloadCid from redis at key:", key, ",with height: ", dagHeight)

	res, err := r.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: height,
		Max: height,
	}).Result()
	if err != nil {
		log.Error("Could not Get payload cid from redis error: ", err)

		return "", err
	}

	log.Debug("Result for ZRangeByScoreWithScores : ", res)

	if len(res) == 1 {
		payloadCid = fmt.Sprintf("%v", res[0].Member)
	}

	log.Debugf("Geted %d Payload CIDs for key %s", len(res), key)

	return payloadCid, nil
}

// GetLastVerifiedDagHeight returns the last verified dag block height for the project
func (r *RedisCache) GetLastVerifiedDagHeight(ctx context.Context, projectID string) (int, error) {
	log.WithField("projectID", projectID)
	key := fmt.Sprintf(redisutils.REDIS_KEY_DAG_LAST_VERIFIED_HEIGHT, projectID)

	val, err := r.redisClient.HGet(ctx, key, projectID).Result()
	if err != nil {
		if err == redis.Nil {
			log.Errorf("error getting verification status, key %s not found", key)

			return 0, nil
		}

		return 0, ErrGettingLastDagVerificationStatus
	}

	height, err := strconv.Atoi(val)
	if err != nil {
		log.Errorf("error converting last verified height to int")

		return 0, ErrGettingLastDagVerificationStatus
	}

	return height, nil
}

func (r *RedisCache) UpdateDagVerificationStatus(ctx context.Context, projectID string, status *datamodel.DagVerifierStatus) error {
	log.WithField("projectID", projectID)
	key := fmt.Sprintf(redisutils.REDIS_KEY_DAG_VERIFICATION_STATUS, projectID)

	data, err := json.Marshal(status)
	if err != nil {
		log.Errorf("error marshalling verification status for project %s", projectID)

		return err
	}

	_, err = r.redisClient.HSet(ctx, key, status.Timestamp, string(data)).Result()
	if err != nil {
		log.Errorf("error updating last verified dag height for project %s", projectID)

		return err
	}

	return nil
}

func (r *RedisCache) GetProjectDAGBlockHeight(ctx context.Context, projectID string) (int, error) {
	//TODO implement me
	panic("implement me")
}

// UpdateDAGChainIssues updates the gaps in the dag chain for the project
func (r *RedisCache) UpdateDAGChainIssues(ctx context.Context, projectID string, dagChainIssues []*datamodel.DagChainIssue) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_DAG_CHAIN_GAPS, projectID)
	gaps := make([]*redis.Z, 0)

	l := log.WithField("projectID", projectID).WithField("dagChainIssues", dagChainIssues)

	for i := range dagChainIssues {
		gapStr, err := json.Marshal(dagChainIssues[i])
		if err != nil {
			l.WithError(err).Error("CRITICAL: failed to marshal dagChainIssue into json")
			continue
		}

		gaps = append(gaps, &redis.Z{Score: float64(dagChainIssues[i].DAGBlockHeight), Member: gapStr})
	}

	_, err := r.redisClient.ZAdd(ctx, key, gaps...).Result()
	if err != nil {
		l.WithError(err).Error("failed to update dagChainGaps into redis")

		return err
	}

	log.Infof("added %d DagGaps data successfully in redis for project: %s", len(dagChainIssues), projectID)

	return nil
}

// GetPayloadCIDs return the list of payloads cids for the project in given range
// startHeight and endHeight are string because they can be "-inf" or "+inf"
// -inf & +inf are just alias for start and end respectively, though the values must be changed according to cache implementation
func (r *RedisCache) GetPayloadCIDs(ctx context.Context, projectID string, startHeight, endHeight string) ([]*datamodel.DagBlock, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID)

	log.
		WithField("key", key).
		Debug("fetching payload CIDs from redis")

	val, err := r.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: startHeight,
		Max: endHeight,
	}).Result()

	if err != nil {
		log.Error("Could not fetch payload CIDs error: ", err)

		return nil, err
	}

	dagChainBlocks := make([]*datamodel.DagBlock, len(val))

	for index, entry := range val {
		dagChainBlocks[index] = &datamodel.DagBlock{
			Data:   &datamodel.Data{PayloadLink: &datamodel.IPLDLink{Cid: fmt.Sprintf("%s", entry.Member)}},
			Height: int64(entry.Score),
		}
	}

	return dagChainBlocks, nil
}

// GetDagChainCIDs return the list of dag chain cids for the project in given range
// startHeight and endHeight are string because they can be "-inf" or "+inf"
// -inf & +inf are just alias for start and end respectively, though the values must be changed according to cache implementation
func (r *RedisCache) GetDagChainCIDs(ctx context.Context, projectID string, startHeight, endHeight string) ([]*datamodel.DagBlock, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectID)

	val, err := r.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: startHeight,
		Max: endHeight,
	}).Result()

	if err != nil {
		log.Error("Could not fetch entries error: ", err)

		return nil, err
	}

	dagChainBlocks := make([]*datamodel.DagBlock, len(val))

	for index, v := range val {
		dagChainBlocks[index] = &datamodel.DagBlock{
			CurrentCid: v.Member.(string),
			Height:     int64(v.Score),
		}
	}

	return dagChainBlocks, nil
}

// GetStoredProjects returns the list of projects that are stored in redis
func (r *RedisCache) GetStoredProjects(ctx context.Context) ([]string, error) {
	key := redisutils.REDIS_KEY_STORED_PROJECTS
	val, err := r.redisClient.SMembers(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Errorf("error getting stored projects, key %s not found", key)

			return []string{}, nil
		}

		return nil, ErrGettingProjects
	}

	return val, nil
}

func (r *RedisCache) StorePruningIssueReport(ctx context.Context, report *datamodel.PruningIssueReport) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisCache) GetPruningVerificationStatus(ctx context.Context) (map[string]*datamodel.ProjectPruningVerificationStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisCache) UpdatePruningVerificationStatus(ctx context.Context, projectID string, status *datamodel.ProjectPruningVerificationStatus) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisCache) GetProjectDagSegments(ctx context.Context, projectID string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisCache) StoreReportedIssues(ctx context.Context, issue *datamodel.IssueReport) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisCache) RemoveOlderReportedIssues(ctx context.Context, tillTime int) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisCache) GetReportedIssues(ctx context.Context, projectID string) ([]*datamodel.IssueReport, error) {
	//TODO implement me
	panic("implement me")
}

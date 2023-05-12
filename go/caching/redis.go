package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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

func (r *RedisCache) GetSnapshotCidAtEpochID(ctx context.Context, projectID string, epochId int) (string, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID)
	snapshotCid := ""
	height := strconv.Itoa(epochId)

	log.WithField("projectID", projectID).
		WithField("epochId", epochId).
		WithField("key", key).
		Debug("getting snapshotCid from redis at given epochId from the given projectId")

	res, err := r.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: height,
		Max: height,
	}).Result()
	if err != nil {
		log.Error("could not Get snapshot cid from redis error: ", err)

		return "", err
	}

	if len(res) == 1 {
		p := new(datamodel.UnfinalizedSnapshot)

		val, ok := res[0].Member.(string)
		if !ok {
			log.Error("CRITICAL: could not convert snapshot cid data stored in redis to string")
		}

		err = json.Unmarshal([]byte(val), p)
		if err != nil {
			log.WithError(err).Error("CRITICAL: could not unmarshal snapshot cid data stored in redis")

			return "", err
		}

		snapshotCid = p.SnapshotCID

		log.WithField("snapshotCid", snapshotCid).WithField("epochId", epochId).Debug("got snapshot cid at given epochId")
	}

	return snapshotCid, nil
}

// GetStoredProjects returns the list of projects that are stored in redis.
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

func (r *RedisCache) CheckIfProjectExists(ctx context.Context, projectID string) (bool, error) {
	res, err := r.redisClient.Keys(ctx, fmt.Sprintf("projectID:%s:*", projectID)).Result()
	if err != nil {
		log.WithError(err).Error("failed to check if project exists")

		return false, err
	}

	if len(res) == 0 {
		return false, nil
	}

	return true, nil
}

// StoreProjects stores the given projects in the redis cache.
func (r *RedisCache) StoreProjects(background context.Context, projects []string) error {
	_, err := r.redisClient.SAdd(background, redisutils.REDIS_KEY_STORED_PROJECTS, projects).Result()

	if err != nil {
		log.WithError(err).Error("failed to store projects")

		return err
	}

	return nil
}

// AddUnfinalizedSnapshotCID adds the given snapshot cid to the given project's zset.
func (r *RedisCache) AddUnfinalizedSnapshotCID(ctx context.Context, projectID, snapshotCid string, height float64) error {
	p := new(datamodel.UnfinalizedSnapshot)

	p.Expiration = time.Now().Unix() + 3600*24 // 1 day
	p.SnapshotCID = snapshotCid

	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID)

	data, _ := json.Marshal(p)

	err := r.redisClient.ZAdd(ctx, key, &redis.Z{
		Score:  height,
		Member: string(data),
	}).Err()
	if err != nil {
		log.WithError(err).Error("failed to add snapshot cid to zset")

		return err
	}

	// get all the members
	res, err := r.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		log.WithError(err).Error("failed to get all members from zset")

		// ignore error
		return nil
	}

	// remove all the members that have expired
	for _, member := range res {
		m := new(datamodel.UnfinalizedSnapshot)

		val, ok := member.Member.(string)
		if !ok {
			log.Error("CRITICAL: could not convert snapshot cid data stored in redis to string")
		}

		err = json.Unmarshal([]byte(val), m)
		if err != nil {
			log.WithError(err).Error("CRITICAL: could not unmarshal snapshot cid data stored in redis")

			continue
		}

		if float64(m.Expiration) < float64(time.Now().Unix()) {
			err = r.redisClient.ZRem(ctx, key, member.Member).Err()
			if err != nil {
				log.WithError(err).Error("failed to remove expired snapshot cid from zset")
			}
		}
	}

	log.WithField("projectID", projectID).
		WithField("snapshotCid", snapshotCid).
		WithField("epochId", height).
		Debug("added snapshot CID to zset")

	return nil
}

// AddSnapshotterStatusReport adds the snapshotter's status report to the given project and epoch ID.
func (r *RedisCache) AddSnapshotterStatusReport(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_SNAPSHOTTER_STATUS_REPORT, projectId)

	reportJson, err := json.Marshal(report)
	if err != nil {
		log.WithError(err).Error("failed to marshal snapshotter status report")

		return err
	}

	err = r.redisClient.HSet(ctx, key, strconv.Itoa(epochId), string(reportJson)).Err()
	if err != nil {
		log.WithError(err).Error("failed to add snapshotter status report in redis")

		return err
	}

	if report.Missed {
		key = fmt.Sprintf(redisutils.REDIS_KEY_TOTAL_MISSED_SNAPSHOT_COUNT, projectId)
	} else {
		key = fmt.Sprintf(redisutils.REDIS_KEY_TOTAL_SUCCESSFUL_SNAPSHOT_COUNT, projectId)
	}

	err = r.redisClient.Incr(ctx, key).Err()
	if err != nil {
		log.WithError(err).Error("failed to increment total missed snapshot count")
	}

	log.Debug("added snapshotter status report in redis")

	return nil
}

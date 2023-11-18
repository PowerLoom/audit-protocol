package caching

import (
	"context"
	"encoding/json"
	"errors"
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
	readClient  *redis.Client
	writeClient *redis.Client
}

var _ DbCache = (*RedisCache)(nil)

func NewRedisCache(readClient, writeClient *redis.Client) *RedisCache {
	cache := &RedisCache{readClient: readClient, writeClient: writeClient}

	if err := gi.Inject(cache); err != nil {
		log.Fatal("Failed to inject redis cache", err)
	}

	return cache
}

func (r *RedisCache) UpdateEpochProcessingStatus(ctx context.Context, projectID string, epochId int, status string, err string) error {
	htable := fmt.Sprintf(redisutils.REDIS_KEY_EPOCH_STATE_ID, strconv.Itoa(epochId), datamodel.RELAYER_SEND_STATE_ID)

	l := log.WithField("hash table", htable).
		WithField("projectID", projectID).WithField("epochId", epochId).WithField("status", status)

	transition_status_item := datamodel.SnapshotterStateUpdate{
		Status:    status,
		Error:     err,
		Timestamp: time.Now().Unix(),
	}
	state_update_bytes, err2 := json.Marshal(transition_status_item)

	if err2 != nil {
		log.WithError(err2).Error("failed to marshal state update message on relayer submission")
		return err2
	}
	err_ := r.writeClient.HSet(
		ctx,
		htable,
		projectID,
		string(state_update_bytes),
	).Err()
	if err_ != nil {
		l.WithError(err_).Error("failed to update epoch processing status in redis")
		return err_
	}
	return nil
}

func (r *RedisCache) GetUnfinalizedSnapshotAtEpochID(ctx context.Context, projectID string, epochId int) (*datamodel.UnfinalizedSnapshot, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID)
	height := strconv.Itoa(epochId)

	l := log.WithField("projectID", projectID).
		WithField("epochId", epochId).
		WithField("key", key)
	l.Debug("getting snapshotCid from redis at given epochId from the given projectId")

	res, err := r.readClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: height,
		Max: height,
	}).Result()
	l.WithField("redis result ", res).Debug("Got result from redis")
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		l.Error("could not Get snapshot cid from redis error: ", err)

		return nil, err
	}

	if len(res) == 1 {
		p := new(datamodel.UnfinalizedSnapshot)

		val, ok := res[0].Member.(string)
		if !ok {
			l.Error("CRITICAL: could not convert snapshot cid data stored in redis to string")
		}

		err = json.Unmarshal([]byte(val), p)
		if err != nil {
			l.WithError(err).Error("CRITICAL: could not unmarshal snapshot cid data stored in redis")

			return nil, err
		}

		l.WithField("snapshotCid", p.SnapshotCID).WithField("epochId", epochId).Debug("got snapshot at given epochId")

		return p, nil
	}

	return nil, errors.New("could not get snapshot cid at given epochId")
}

// GetStoredProjects returns the list of projects that are stored in redis.
func (r *RedisCache) GetStoredProjects(ctx context.Context) ([]string, error) {
	key := redisutils.REDIS_KEY_STORED_PROJECTS

	val, err := r.readClient.SMembers(ctx, key).Result()
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
	res, err := r.readClient.Keys(ctx, fmt.Sprintf("projectID:%s:*", projectID)).Result()
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
	_, err := r.writeClient.SAdd(background, redisutils.REDIS_KEY_STORED_PROJECTS, projects).Result()

	if err != nil {
		log.WithError(err).Error("failed to store projects")

		return err
	}

	return nil
}

// AddSnapshotterStatusReport adds the snapshotter's status report to the given project and epoch ID.
func (r *RedisCache) AddSnapshotterStatusReport(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_SNAPSHOTTER_STATUS_REPORT, projectId)
	if report != nil {
		reportJson, err := json.Marshal(report)
		if err != nil {
			log.WithError(err).Error("failed to marshal snapshotter status report")

			return err
		}

		err = r.writeClient.HSet(ctx, key, strconv.Itoa(epochId), string(reportJson)).Err()
		if err != nil {
			log.WithError(err).Error("failed to add snapshotter status report in redis")

			return err
		}

		incrKey := ""
		if report.State == datamodel.SuccessfulSnapshotSubmission {
			incrKey = redisutils.REDIS_KEY_TOTAL_SUCCESSFUL_SNAPSHOT_COUNT
		} else if report.State == datamodel.IncorrectSnapshotSubmission {
			incrKey = redisutils.REDIS_KEY_TOTAL_INCORRECT_SNAPSHOT_COUNT
		} else if report.State == datamodel.MissedSnapshotSubmission {
			incrKey = redisutils.REDIS_KEY_TOTAL_MISSED_SNAPSHOT_COUNT
		}
		if incrKey != "" {
			err = r.writeClient.Incr(ctx, fmt.Sprintf(incrKey, projectId)).Err()
			if err != nil {
				log.WithError(err).Error("failed to increment snapshotter status report count in redis")
			}
		}

	}

	log.Debug("added snapshotter status report in redis")

	return nil
}

func (r *RedisCache) StoreLastFinalizedEpoch(ctx context.Context, projectID string, epochId int) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_LAST_FINALIZED_EPOCH, projectID)

	l := log.WithField("key", key)

	lastEpochStr, err := r.readClient.Get(ctx, key).Result()
	// if err != nil {
	// 	l.WithError(err).Error("failed to get last finalized epoch from redis")
	// }

	lastEpoch := 0

	if lastEpochStr != "" {
		lastEpoch, err = strconv.Atoi(lastEpochStr)
		if err != nil {
			l.WithError(err).Error("failed to convert last finalized epoch to int")
		}
	}

	if epochId <= lastEpoch {
		l.WithField("epochId", epochId).Debug("epochId is less than or equal to last finalized epoch, skipping update")

		return nil
	}

	err = r.writeClient.Set(ctx, key, epochId, 0).Err()
	if err != nil {
		l.WithError(err).Error("failed to store last finalized epoch in redis")

		return err
	}

	l.Debug("stored last finalized epoch in redis")

	return nil
}

func (r *RedisCache) StoreFinalizedSnapshot(ctx context.Context, msg *datamodel.PowerloomSnapshotFinalizedMessage) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_FINALIZED_SNAPSHOTS, msg.ProjectID)
	l := log.WithField("msg", msg)

	msg.Expiry = msg.Timestamp + 3600*24 // 1 day

	err := r.writeClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(msg.EpochID),
		Member: msg.SnapshotCID,
	}).Err()

	if err != nil {
		l.WithError(err).Error("failed to add finalized snapshot to zset")

		return err
	}

	// get all the members
	res, err := r.writeClient.ZRemRangeByScore(ctx, key, "-inf", strconv.Itoa(msg.EpochID-10240)).Result()
	if err != nil {
		log.WithField("epoch ID", msg.EpochID).WithField("Redis Op result:", res).WithError(err).Error("failed to remove finalized snapshots zset cache between -inf to epochId-10240")
		// ignore error
		return nil
	}
	log.WithField("epoch ID", msg.EpochID).WithField("Redis Op result:", res).Debug("removed finalized snapshots zset cache between -inf to epochId-10240")
	return nil
}

func (r *RedisCache) GetFinalizedSnapshotAtEpochID(ctx context.Context, projectID string, epochId int) (string, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_FINALIZED_SNAPSHOTS, projectID)
	l := log.WithField("projectId", projectID).WithField("epochId", epochId)

	res, err := r.readClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: strconv.Itoa(epochId),
		Max: strconv.Itoa(epochId),
	}).Result()
	if err != nil {
		l.WithError(err).Error("failed to get finalized snapshot from zset")

		return "", err
	}

	if len(res) == 0 {
		l.Debug("no finalized snapshot found at epochId")

		return "", fmt.Errorf("no finalized snapshot found at epochId %d", epochId)
	}

	if len(res) > 1 {
		l.Debug("more than one finalized snapshot found at epochId")

		return "", fmt.Errorf("more than one finalized snapshot found at epochId %d", epochId)
	}

	snapshot_cid, ok := res[0].Member.(string)
	if !ok {
		l.Error("CRITICAL: could not convert finalized snapshot data stored in redis to string")

		return "", errors.New("failed to convert finalized snapshot data stored in redis to string")
	}

	return snapshot_cid, nil
}

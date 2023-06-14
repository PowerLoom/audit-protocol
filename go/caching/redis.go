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

func (r *RedisCache) GetUnfinalizedSnapshotAtEpochID(ctx context.Context, projectID string, epochId int) (*datamodel.UnfinalizedSnapshot, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID)
	snapshotCid := ""
	height := strconv.Itoa(epochId)

	log.WithField("projectID", projectID).
		WithField("epochId", epochId).
		WithField("key", key).
		Debug("getting snapshotCid from redis at given epochId from the given projectId")

	res, err := r.readClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: height,
		Max: height,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		log.Error("could not Get snapshot cid from redis error: ", err)

		return nil, err
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

			return nil, err
		}

		log.WithField("snapshotCid", snapshotCid).WithField("epochId", epochId).Debug("got snapshot at given epochId")

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

// AddUnfinalizedSnapshotCID adds the given snapshot cid to the given project's zset.
func (r *RedisCache) AddUnfinalizedSnapshotCID(ctx context.Context, msg *datamodel.PayloadCommitMessage) error {
	p := new(datamodel.UnfinalizedSnapshot)

	p.Expiration = time.Now().Unix() + 3600*24 // 1 day
	p.SnapshotCID = msg.SnapshotCID

	p.Snapshot = msg.Message

	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, msg.ProjectID)

	data, _ := json.Marshal(p)

	err := r.writeClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(msg.EpochID),
		Member: string(data),
	}).Err()
	if err != nil {
		log.WithError(err).Error("failed to add snapshot cid to zset")

		return err
	}

	// get all the members
	res, err := r.writeClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
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
			err = r.writeClient.ZRem(ctx, key, member.Member).Err()
			if err != nil {
				log.WithError(err).Error("failed to remove expired snapshot cid from zset")
			}
		}
	}

	log.WithField("projectID", msg.ProjectID).
		WithField("snapshotCid", msg.SnapshotCID).
		WithField("epochId", msg.EpochID).
		Debug("added snapshot CID to zset")

	return nil
}

// AddSnapshotterStatusReport adds the snapshotter's status report to the given project and epoch ID.
func (r *RedisCache) AddSnapshotterStatusReport(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport, incrCount bool) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_SNAPSHOTTER_STATUS_REPORT, projectId)

	storedReport := new(datamodel.SnapshotterStatusReport)

	reportJsonString, err := r.readClient.HGet(ctx, key, strconv.Itoa(epochId)).Result()
	if err == nil || reportJsonString != "" {
		_ = json.Unmarshal([]byte(reportJsonString), storedReport)
	}

	if report != nil {
		if storedReport.SubmittedSnapshotCid != "" {
			report.SubmittedSnapshotCid = storedReport.SubmittedSnapshotCid
		}

		if storedReport.Reason != "" {
			report.Reason = storedReport.Reason
		}

		if storedReport.FinalizedSnapshotCid != "" {
			report.FinalizedSnapshotCid = storedReport.FinalizedSnapshotCid
		}

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

		switch {
		case report.State == datamodel.MissedSnapshotSubmission:
			key = fmt.Sprintf(redisutils.REDIS_KEY_TOTAL_MISSED_SNAPSHOT_COUNT, projectId)
		case report.State == datamodel.IncorrectSnapshotSubmission:
			key = fmt.Sprintf(redisutils.REDIS_KEY_TOTAL_INCORRECT_SNAPSHOT_COUNT, projectId)
		}
	} else {
		key = fmt.Sprintf(redisutils.REDIS_KEY_TOTAL_SUCCESSFUL_SNAPSHOT_COUNT, projectId)
	}

	if incrCount {
		err = r.writeClient.Incr(ctx, key).Err()
		if err != nil {
			log.WithError(err).Error("failed to increment total missed snapshot count")
		}
	}

	log.Debug("added snapshotter status report in redis")

	return nil
}

func (r *RedisCache) StoreLastFinalizedEpoch(ctx context.Context, projectID string, epochId int) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_LAST_FINALIZED_EPOCH, projectID)

	l := log.WithField("key", key)

	lastEpochStr, err := r.readClient.Get(ctx, key).Result()
	if err != nil {
		l.WithError(err).Error("failed to get last finalized epoch from redis")
	}

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

	l.WithField("projectID", projectID).WithField("epochId", epochId).Debug("stored last finalized epoch in redis")

	return nil
}

func (r *RedisCache) StoreFinalizedSnapshot(ctx context.Context, msg *datamodel.PowerloomSnapshotFinalizedMessage) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_FINALIZED_SNAPSHOTS, msg.ProjectID)
	l := log.WithField("msg", msg)

	msg.Expiry = msg.Timestamp + 3600*24 // 1 day

	data, err := json.Marshal(msg)
	if err != nil {
		l.WithError(err).Error("failed to marshal finalized snapshot")

		return err
	}

	err = r.writeClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(msg.EpochID),
		Member: string(data),
	}).Err()

	if err != nil {
		l.WithError(err).Error("failed to add finalized snapshot to zset")

		return err
	}

	// get all the members
	res, err := r.readClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		log.WithError(err).Error("failed to get all finalized snapshots")

		// ignore error
		return nil
	}

	// remove all the members that have expired
	membersToRemove := make([]interface{}, 0)
	timeNow := time.Now().Unix()

	for _, member := range res {
		m := new(datamodel.PowerloomSnapshotFinalizedMessage)

		val, ok := member.Member.(string)
		if !ok {
			log.Error("CRITICAL: could not convert snapshot cid data stored in redis to string")
		}

		err = json.Unmarshal([]byte(val), m)
		if err != nil {
			log.WithError(err).Error("CRITICAL: could not unmarshal snapshot cid data stored in redis")

			continue
		}

		if float64(m.Expiry) < float64(timeNow) {
			membersToRemove = append(membersToRemove, member.Member)
		}
	}

	if len(membersToRemove) == 0 {
		return nil
	}

	err = r.writeClient.ZRem(ctx, key, membersToRemove).Err()
	if err != nil {
		log.WithError(err).Error("failed to remove expired snapshot cid from zset")
	}

	return nil
}

func (r *RedisCache) GetFinalizedSnapshotAtEpochID(ctx context.Context, projectID string, epochId int) (*datamodel.PowerloomSnapshotFinalizedMessage, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_FINALIZED_SNAPSHOTS, projectID)
	l := log.WithField("projectId", projectID).WithField("epochId", epochId)

	res, err := r.readClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: strconv.Itoa(epochId),
		Max: strconv.Itoa(epochId),
	}).Result()
	if err != nil {
		l.WithError(err).Error("failed to get finalized snapshot from zset")

		return nil, err
	}

	if len(res) == 0 {
		l.Debug("no finalized snapshot found at epochId")

		return nil, fmt.Errorf("no finalized snapshot found at epochId %d", epochId)
	}

	if len(res) > 1 {
		l.Debug("more than one finalized snapshot found at epochId")

		return nil, fmt.Errorf("more than one finalized snapshot found at epochId %d", epochId)
	}

	m := new(datamodel.PowerloomSnapshotFinalizedMessage)

	val, ok := res[0].Member.(string)
	if !ok {
		l.Error("CRITICAL: could not convert finalized snapshot data stored in redis to string")

		return nil, errors.New("failed to convert finalized snapshot data stored in redis to string")
	}

	err = json.Unmarshal([]byte(val), m)
	if err != nil {
		l.WithError(err).Error("CRITICAL: could not unmarshal finalized snapshot data stored in redis")

		return nil, err
	}

	return m, nil
}

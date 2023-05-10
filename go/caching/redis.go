package caching

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

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

func (r *RedisCache) GetPayloadCidAtEpochID(ctx context.Context, projectID string, epochId int) (string, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID)
	payloadCid := ""
	height := strconv.Itoa(epochId)

	log.WithField("projectID", projectID).
		WithField("epochId", epochId).
		WithField("key", key).
		Debug("getting PayloadCid from redis at given epochId from the given projectId")

	res, err := r.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: height,
		Max: height,
	}).Result()
	if err != nil {
		log.Error("could not Get payload cid from redis error: ", err)

		return "", err
	}

	if len(res) == 1 {
		payloadCid = fmt.Sprintf("%v", res[0].Member)
		log.WithField("result", res).WithField("epochId", epochId).Debug("got payload cid at given epochId")
	}

	return payloadCid, nil
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

// StoreProjects stores the given projects in the redis cache
func (r *RedisCache) StoreProjects(background context.Context, projects []string) error {
	_, err := r.redisClient.SAdd(background, redisutils.REDIS_KEY_STORED_PROJECTS, projects).Result()

	if err != nil {
		log.WithError(err).Error("failed to store projects")

		return err
	}

	return nil
}

func (r *RedisCache) AddPayloadCID(ctx context.Context, projectID, payloadCID string, height float64) error {
	pipe := r.getPipelinerFromCtx(ctx)
	var err error

	if pipe != nil {
		err = pipe.ZAdd(ctx, fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID), &redis.Z{
			Score:  height,
			Member: payloadCID,
		}).Err()
	} else {
		err = r.redisClient.ZAdd(ctx, fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID), &redis.Z{
			Score:  height,
			Member: payloadCID,
		}).Err()
	}

	if err != nil {
		log.WithError(err).Error("failed to add payload CID to zset")
	}

	log.WithField("projectID", projectID).
		WithField("payloadCID", payloadCID).
		WithField("epochId", height).
		Debug("added payload CID to zset")

	return err
}

func (r *RedisCache) RemovePayloadCIDAtEpochID(ctx context.Context, projectID string, dagHeight int) error {
	pipe := r.getPipelinerFromCtx(ctx)
	var err error

	if pipe != nil {
		err = pipe.ZRemRangeByScore(ctx, fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID), strconv.Itoa(dagHeight), strconv.Itoa(dagHeight)).Err()
	} else {
		err = r.redisClient.ZRemRangeByScore(ctx, fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID), strconv.Itoa(dagHeight), strconv.Itoa(dagHeight)).Err()
	}

	if err != nil {
		if err == redis.Nil {
			return nil
		}

		log.WithError(err).Error("failed to remove payload CID from zset")
	}

	return err
}

func (r *RedisCache) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	pipe := r.redisClient.TxPipeline()

	ctx = context.WithValue(ctx, "pipeliner", pipe)

	err := fn(ctx)
	if err != nil {
		err = pipe.Discard()
	} else {
		_, err = pipe.Exec(ctx)
		log.WithError(err).Error("failed to execute transaction")
	}

	return err
}

func (r *RedisCache) getPipelinerFromCtx(ctx context.Context) redis.Pipeliner {
	if pipeliner, ok := ctx.Value("pipeliner").(redis.Pipeliner); ok {
		return pipeliner
	}

	return nil
}

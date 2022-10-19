package redisutils

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

func InitRedisClient(redisURL string, redisDb int, poolSize int) *redis.Client {

	log.Info("Connecting to redis at:", redisURL)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       redisDb,
		PoolSize: poolSize,
	})
	pong, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Error("Unable to connect to redis at:")
	}
	log.Info("Connected successfully to Redis and received ", pong, " back")
	return redisClient
}

func FetchStoredProjects(ctx context.Context, redisClient *redis.Client, retryCount int) []string {
	key := REDIS_KEY_STORED_PROJECTS
	log.Debugf("Fetching stored Projects from redis at key: %s", key)
	for i := 0; i < retryCount; i++ {
		res := redisClient.SMembers(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				log.Infof("Stored Projects key doesn't exist..retrying")
				time.Sleep(5 * time.Minute)
				continue
			}
			log.Errorf("Failed to fetch stored projects from redis due to err %+v. Retrying %d", res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Infof("Retrieved %d storedProjects %+v from redis", len(res.Val()), res.Val())
		return res.Val()
	}
	return nil
}

package redisutils

import (
	"context"
	"net"
	"strconv"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

func InitReaderRedisClient(redisHost string, port int, redisDb int, poolSize int, password string) *redis.Client {
	redisURL := net.JoinHostPort(redisHost, strconv.Itoa(port))

	log.Info("connecting to reader redis at:", redisURL)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: password,
		DB:       redisDb,
		PoolSize: poolSize,
	})

	pong, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.WithField("addr", redisURL).Fatal("unable to connect to reader redis")
	}

	log.Info("connected successfully to reader Redis and received ", pong, " back")

	return redisClient
}

func InitWriterRedisClient(redisHost string, port int, redisDb int, poolSize int, password string) *redis.Client {
	redisURL := net.JoinHostPort(redisHost, strconv.Itoa(port))

	log.Info("connecting to writer redis at:", redisURL)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: password,
		DB:       redisDb,
		PoolSize: poolSize,
	})

	pong, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.WithField("addr", redisURL).Fatal("Unable to connect to writer redis")
	}

	log.Info("Connected successfully to writer Redis and received ", pong, " back")

	return redisClient
}

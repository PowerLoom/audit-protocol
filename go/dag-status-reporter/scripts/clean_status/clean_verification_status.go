package main

import (
    "context"

    "audit-protocol/goutils/redisutils"
    "audit-protocol/goutils/settings"
)

func main() {
    obj := settings.ParseSettings()

    client := redisutils.InitRedisClient(obj.Redis.Host, obj.Redis.Port, obj.Redis.Db, obj.DagVerifierSettings.RedisPoolSize, obj.Redis.Password, -1)

    keys, err := client.Keys(context.Background(), "*:lastReportedHeight").Result()
    if err != nil {
        panic(err)
    }

    reportsKeys, err := client.Keys(context.Background(), "*:dagVerificationStatus").Result()
    if err != nil {
        panic(err)
    }

    keys = append(keys, reportsKeys...)

    err = client.Del(context.Background(), keys...).Err()
    if err != nil {
        panic(err)
    }
}

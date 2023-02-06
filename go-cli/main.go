package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/powerloom/audit-prototol-private/goutils/settings"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var settingsObj *settings.SettingsObj
var redisClient *redis.Client

func main() {
	cmd := flag.Int("c", 1, "command to be executed, (1)populate-redis-zsets")
	fileName := flag.String("file", "", "file name to be used as input")
	flag.Parse()

	settingsObj = settings.ParseSettings("../settings.json")
	InitRedisClient()
	if *cmd == 1 {
		if fileName == nil || *fileName == "" {
			fmt.Println("Empty filename passed with command (1)populate-redis-zsets")
			flag.PrintDefaults()
		} else {
			PopulateZSets(*fileName)
		}
	} else {
		fmt.Printf("Command passed is not in the list of supported commands. Supported commands are : (1) populate-redis-zsets")
	}
}

const REDIS_KEY_PROJECT_PAYLOAD_CIDS string = "projectID:%s:payloadCids"
const REDIS_KEY_PROJECT_CIDS string = "projectID:%s:Cids"

type ZSets struct {
	PayloadCids *map[string]string `json:"payloadCids"`
	DagCids     *map[string]string `json:"dagCids"`
}

func PopulateZSets(fileName string) {
	//uniswap_pairContract_pair_total_reserves_0x6591c4bcd6d7a1eb4e537da8b78676c1576ba244_UNISWAPV2__0_5000.json
	strs := strings.Split(fileName, "__")
	projectId := strs[0]
	fileName = settingsObj.PruningServiceSettings.CARStoragePath + fileName
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Unable to file %s due to error %+v\n", fileName, err)
		return
	}
	defer file.Close()
	var zsets ZSets
	byteValue, _ := ioutil.ReadAll(file)
	err = json.Unmarshal(byteValue, &zsets)
	if err != nil {
		fmt.Printf("Failed to unmarshal json struct from file %s due to errro %+v\n", fileName, err)
		return
	}
	fmt.Printf("%d payloadCids present in file %s\n", len(*zsets.PayloadCids), fileName)
	fmt.Printf("%d Cids present in file %s\n", len(*zsets.DagCids), fileName)
	fmt.Println("payloadCids", zsets.PayloadCids)
	key := fmt.Sprintf(REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	PopulateZSet(key, zsets.PayloadCids)
	key = fmt.Sprintf(REDIS_KEY_PROJECT_CIDS, projectId)
	PopulateZSet(key, zsets.DagCids)

}

func PopulateZSet(key string, cids *map[string]string) {
	length := len(*cids)
	fmt.Println("Redis key :", key, ", length:", length)
	var scoresMembers []*redis.Z
	//fmt.Println("scoresMembers:", scoresMembers)

	for score, value := range *cids {
		//fmt.Println("score :", score)
		intScore, err := strconv.ParseInt(score, 10, 32)
		if err != nil {
			fmt.Print("Error parsing score from file.", err)
			return
		}
		scoreMember := redis.Z{Score: float64(intScore), Member: value}
		//fmt.Println("scoreMember :", scoreMember)
		scoresMembers = append(scoresMembers, &scoreMember)
	}
	//fmt.Println("scoresMembers:", scoresMembers)
	for i := 0; ; i++ {
		res := redisClient.ZAddNX(ctx, key, scoresMembers...)
		if res.Err() != nil {
			fmt.Printf("Could not add Cids to redis key %s due to error %+v. Retrying %d.\n", key, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Printf("Added %d payload Cids to redis key %s\n", length, key)
		return
	}
}

func InitRedisClient() {
	redisURL := fmt.Sprintf("%s:%d", settingsObj.Redis.Host, settingsObj.Redis.Port)
	redisDb := settingsObj.Redis.Db
	fmt.Printf("Connecting to redis DB %d at %s\n", redisDb, redisURL)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: settingsObj.Redis.Password,
		DB:       redisDb,
	})
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Unable to connect to redis at %s with error %+v\n", redisURL, err)
	}
	fmt.Println("Connected successfully to Redis and received ", pong, " back")
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/remeh/sizedwaitgroup"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/redisutils"
)

type migrator struct {
	follower   *redis.Client
	master     *redis.Client
	ipfsclient *ipfsutils.IpfsClient
}

func main() {
	// make sure you have two connections to the different redis instances
	follower := redisutils.InitRedisClient("localhost", 6379, 0, 100, "", -1)
	defer follower.Close()

	master := redisutils.InitRedisClient("localhost", 6380, 0, 100, "", -1)
	defer master.Close()

	ipfsclient := ipfsutils.InitClient("localhost:5001", 100, nil, 10)

	m := &migrator{
		follower:   follower,
		master:     master,
		ipfsclient: ipfsclient,
	}

	m.fill()
}

func (m *migrator) fill() {
	log.Println("filling follower with data from master")

	patterns := []string{"project*", "storedProjectIds"}

	swg := sizedwaitgroup.New(len(patterns))
	for _, pattern := range patterns {
		swg.Add()

		go func(pattern string) {
			defer swg.Done()

			m.dumpAndRestore(pattern)
		}(pattern)
	}

	swg.Wait()

	// keep only two projects in the follower for testing purpose
	projects, err := m.follower.SMembers(context.Background(), redisutils.REDIS_KEY_STORED_PROJECTS).Result()
	if err != nil {
		log.Fatalln("error getting stored projects", err)
	}

	keysToDelete := make([]string, 0)

	for _, project := range projects[2:] {
		// delete the project from the follower
		err = m.follower.SRem(context.Background(), redisutils.REDIS_KEY_STORED_PROJECTS, project).Err()

		keys, err := m.follower.Keys(context.Background(), fmt.Sprintf("projectID:%s:*", project)).Result()
		if err != nil {
			log.Fatalln("error getting keys for pattern project*", err)
		}

		keysToDelete = append(keysToDelete, keys...)
	}

	err = m.follower.HDel(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PRUNING_STATUS), projects[2:]...).Err()
	if err != nil {
		log.Fatalln("error deleting pruning status", err)
	}

	// delete the keys
	err = m.follower.Del(context.Background(), keysToDelete...).Err()
	if err != nil {
		log.Fatalln("error deleting keys", err)
	}

	// fill dummy data
	projectIDs := projects[:2]
	if err != nil {
		log.Fatalln("error getting stored projects", err)
	}

	for _, projectID := range projectIDs {
		// replacing dag segment CIDs with dummy CIDs so that we can test the archival flow
		segmentMap, err := m.follower.HGetAll(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectID)).Result()
		if err != nil {
			log.Fatalln("error getting project metadata", err)
		}

		segment := make(map[string]interface{})

		// updating dag segments with dummy CIDs
		for key, segmentString := range segmentMap {
			cid := m.createDummyDagCID([]byte(key))

			json.Unmarshal([]byte(segmentString), &segment)

			segment["endDAGCID"] = cid

			data, _ := json.Marshal(segment)

			err = m.follower.HSet(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectID), key, string(data)).Err()
			if err != nil {
				log.Fatalln("error setting project metadata", err)
			}
		}

		// for each project cids and payload cids, replace the CIDs with dummy CIDs
		cids, err := m.follower.ZRangeWithScores(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectID), 0, -1).Result()
		if err != nil {
			log.Fatalln("error getting project cids", err)
		}

		if len(cids) == 0 {
			continue
		}

		payloadCIDs, err := m.follower.ZRangeWithScores(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID), 0, -1).Result()
		if err != nil {
			log.Fatalln("error getting project payload cids", err)
		}

		if len(payloadCIDs) == 0 {
			continue
		}

		err = m.follower.ZRemRangeByScore(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectID), strconv.Itoa(int(cids[0].Score)), strconv.Itoa(int(cids[len(cids)-1].Score))).Err()
		if err != nil {
			log.Fatalln("error removing cids from Cids zset ", err)
		}

		err = m.follower.ZRemRangeByScore(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID), strconv.Itoa(int(payloadCIDs[0].Score)), strconv.Itoa(int(payloadCIDs[len(payloadCIDs)-1].Score))).Err()
		if err != nil {
			log.Fatalln("error removing cids from PayloadCids zset ", err)
		}

		cidMembers := make([]*redis.Z, len(cids))

		for index, cid := range cids {
			dummyCID := m.createDummyDagCID([]byte(projectID + strconv.Itoa(int(cid.Score))))

			cidMembers[index] = &redis.Z{Score: cid.Score, Member: dummyCID}

			if index%500 == 0 {
				time.Sleep(1 * time.Second)
			}
		}

		err = m.follower.ZAdd(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectID), cidMembers...).Err()
		if err != nil {
			log.Fatalln("error adding cid to Cids zset ", err)
		}

		payloadCIDMembers := make([]*redis.Z, len(payloadCIDs))

		for index, cid := range payloadCIDs {
			dummyCID := m.createDummyDagCID([]byte(projectID + "pcid" + strconv.Itoa(int(cid.Score))))

			payloadCIDMembers[index] = &redis.Z{Score: cid.Score, Member: dummyCID}

			if index%500 == 0 {
				time.Sleep(1 * time.Second)
			}
		}

		err = m.follower.ZAdd(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID), payloadCIDMembers...).Err()
		if err != nil {
			log.Fatalln("error adding cid to PayloadCids zset ", err)
		}
	}
}

func (m *migrator) dumpAndRestore(pattern string) {
	val, err := m.master.Keys(context.Background(), pattern).Result()
	if err != nil {
		log.Fatalln("error getting keys for pattern project*", err)
	}

	swg := sizedwaitgroup.New(100)
	for _, key := range val {
		swg.Add()

		go func(key string) {
			defer swg.Done()

			resp, err := m.master.Dump(context.Background(), key).Result()
			if err != nil {
				log.Fatalln("error dumping key", key, err)
			}

			// restore the key in the follower
			err = m.follower.RestoreReplace(context.Background(), key, 0, resp).Err()
			if err != nil {
				log.Fatalln("error restoring key", key, err)
			}
		}(key)
	}

	swg.Wait()
}

// createDummyDagCID creates a dummy dag cid in local node and stores it in the redis instance for testing and simulation purposes.
func (m *migrator) createDummyDagCID(data []byte) string {
	cid, err := m.ipfsclient.AddFileToIPFS(data)
	if err != nil {
		log.Fatalln("error adding file to ipfs", err)
	}

	log.Println("dummy dag cid", cid)

	return cid
}

package caching

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/redisutils"
)

func TestNewRedisCache(t *testing.T) {
	db, _ := redismock.NewClientMock()

	want := &RedisCache{readClient: db, writeClient: db}

	t.Run("successful init", func(t *testing.T) {
		if got := NewRedisCache(db, db); !reflect.DeepEqual(got, want) {
			t.Errorf("NewRedisCache() = %v, want %v", got, want)
		}
	})
}

func TestRedisCache_GetSnapshotAtEpochID(t *testing.T) {
	// Create a new mock client and a corresponding RedisCache instance
	mockClient, mock := redismock.NewClientMock()

	cache := RedisCache{
		readClient:  mockClient,
		writeClient: mockClient,
	}

	projectID := "testProject"
	epochID := 42

	t.Run("Get snapshot at existing epoch ID", func(t *testing.T) {
		expectedSnapshot := &datamodel.UnfinalizedSnapshot{
			SnapshotCID: "snapshotcid",
			Snapshot: map[string]interface{}{
				"dummy": "data",
			},
			TTL: time.Now().Unix(),
		}

		snapshotJSON, _ := json.Marshal(expectedSnapshot)

		mock.ExpectZRangeByScoreWithScores(fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID), &redis.ZRangeBy{
			Min: strconv.Itoa(epochID),
			Max: strconv.Itoa(epochID),
		}).SetVal([]redis.Z{{Member: string(snapshotJSON)}})

		snapshot, err := cache.GetSnapshotAtEpochID(context.Background(), projectID, epochID)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if snapshot == nil {
			t.Errorf("Expected snapshot, got nil")
		} else if snapshot.SnapshotCID != expectedSnapshot.SnapshotCID || !reflect.DeepEqual(snapshot.Snapshot, expectedSnapshot.Snapshot) {
			t.Errorf("Snapshot does not match expected values")
		}
	})

	t.Run("Get snapshot at non-existing epoch ID", func(t *testing.T) {
		mock.ExpectZRangeByScoreWithScores(fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID), &redis.ZRangeBy{
			Min: strconv.Itoa(epochID),
			Max: strconv.Itoa(epochID),
		}).SetVal([]redis.Z{})

		snapshot, err := cache.GetSnapshotAtEpochID(context.Background(), projectID, epochID)
		if err == nil {
			t.Errorf("Expected error, got no error")
		}

		if snapshot != nil {
			t.Errorf("Expected nil snapshot, got: %+v", snapshot)
		}
	})

	t.Run("Redis error - key not found", func(t *testing.T) {
		mock.ExpectZRangeByScoreWithScores(fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID), &redis.ZRangeBy{
			Min: strconv.Itoa(epochID),
			Max: strconv.Itoa(epochID),
		}).RedisNil()

		snapshot, err := cache.GetSnapshotAtEpochID(context.Background(), projectID, epochID)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if snapshot != nil {
			t.Errorf("Expected nil snapshot, got: %+v", snapshot)
		}
	})

	t.Run("Redis error - other error", func(t *testing.T) {
		expectedErr := errors.New("Redis error")

		mock.ExpectZRangeByScoreWithScores(fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID), &redis.ZRangeBy{
			Min: strconv.Itoa(epochID),
			Max: strconv.Itoa(epochID),
		}).SetErr(expectedErr)

		snapshot, err := cache.GetSnapshotAtEpochID(context.Background(), projectID, epochID)
		if err == nil {
			t.Errorf("Expected an error, but no error occurred")
		}

		if snapshot != nil {
			t.Errorf("Expected nil snapshot, got: %+v", snapshot)
		}
	})

	// Ensure that all expectations were met
	err := mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet Redis expectations: %v", err)
	}
}

func TestRedisCache_GetStoredProjects(t *testing.T) {
	mockClient, mock := redismock.NewClientMock()

	cache := RedisCache{
		readClient:  mockClient,
		writeClient: mockClient,
	}

	t.Run("Get stored projects", func(t *testing.T) {
		expectedProjects := []string{"project1", "project2"}

		mock.ExpectSMembers(fmt.Sprintf(redisutils.REDIS_KEY_STORED_PROJECTS)).SetVal(expectedProjects)

		projects, err := cache.GetStoredProjects(context.Background())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !reflect.DeepEqual(projects, expectedProjects) {
			t.Errorf("Projects do not match expected values")
		}
	})

	t.Run("Redis error - key not found", func(t *testing.T) {
		mock.ExpectSMembers(fmt.Sprintf(redisutils.REDIS_KEY_STORED_PROJECTS)).RedisNil()

		projects, err := cache.GetStoredProjects(context.Background())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if projects == nil {
			t.Errorf("Expected empty projects, got nil")
		}
	})

	t.Run("Redis error - other error", func(t *testing.T) {
		expectedErr := errors.New("Redis error")

		mock.ExpectSMembers(fmt.Sprintf(redisutils.REDIS_KEY_STORED_PROJECTS)).SetErr(expectedErr)

		projects, err := cache.GetStoredProjects(context.Background())
		if err == nil {
			t.Errorf("Expected an error, but no error occurred")
		}

		if projects != nil {
			t.Errorf("Expected nil projects, got: %+v", projects)
		}
	})

	// Ensure that all expectations were met
	err := mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet Redis expectations: %v", err)
	}
}

func TestRedisCache_CheckIfProjectExists(t *testing.T) {
	mockClient, mock := redismock.NewClientMock()
	cache := RedisCache{
		readClient:  mockClient,
		writeClient: mockClient,
	}

	projectID := "testProject"

	t.Run("Project exists", func(t *testing.T) {
		mock.ExpectKeys(fmt.Sprintf("projectID:%s:*", projectID)).SetVal([]string{"projectID:testProject:1", "projectID:testProject:2"})

		exists, err := cache.CheckIfProjectExists(context.Background(), projectID)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !exists {
			t.Errorf("Expected project to exist, but it does not exist")
		}
	})

	t.Run("Project does not exist", func(t *testing.T) {
		mock.ExpectKeys(fmt.Sprintf("projectID:%s:*", projectID)).SetVal([]string{})

		exists, err := cache.CheckIfProjectExists(context.Background(), projectID)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if exists {
			t.Errorf("Expected project to not exist, but it exists")
		}
	})

	t.Run("Redis error", func(t *testing.T) {
		expectedErr := errors.New("Redis error")

		mock.ExpectKeys(fmt.Sprintf("projectID:%s:*", projectID)).SetErr(expectedErr)

		exists, err := cache.CheckIfProjectExists(context.Background(), projectID)
		if err == nil {
			t.Errorf("Expected an error, but no error occurred")
		}

		if exists {
			t.Errorf("Expected project to not exist, but it exists")
		}
	})

	// Ensure that all expectations were met
	err := mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet Redis expectations: %v", err)
	}
}

func TestRedisCache_StoreProjects(t *testing.T) {
	mockClient, mock := redismock.NewClientMock()
	cache := RedisCache{
		readClient:  mockClient,
		writeClient: mockClient,
	}

	background := context.Background()
	projects := []string{"project1", "project2", "project3"}

	t.Run("Store projects successfully", func(t *testing.T) {
		mock.ExpectSAdd(redisutils.REDIS_KEY_STORED_PROJECTS, projects).SetVal(3)

		err := cache.StoreProjects(background, projects)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Redis error", func(t *testing.T) {
		expectedErr := errors.New("Redis error")

		mock.ExpectSAdd(redisutils.REDIS_KEY_STORED_PROJECTS, projects).SetErr(expectedErr)

		err := cache.StoreProjects(background, projects)
		if err == nil {
			t.Errorf("Expected an error, but no error occurred")
		}
	})

	// Ensure that all expectations were met
	err := mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet Redis expectations: %v", err)
	}
}

func TestRedisCache_AddUnfinalizedSnapshotCID(t *testing.T) {
	mockClient, mock := redismock.NewClientMock()
	cache := RedisCache{
		readClient:  mockClient,
		writeClient: mockClient,
	}

	ctx := context.Background()
	projectID := "testProject"
	snapshotCID := "snapshotCID"
	epochID := 1

	payload := &datamodel.PayloadCommitMessage{
		ProjectID:   projectID,
		SnapshotCID: snapshotCID,
		EpochID:     epochID,
		Message: map[string]interface{}{
			"dummy": "data",
		},
	}

	t.Run("Add snapshot CID successfully", func(t *testing.T) {
		ttl := time.Now().Unix() + 3600*24
		expectedSnapshot := &datamodel.UnfinalizedSnapshot{
			SnapshotCID: snapshotCID,
			Snapshot:    payload.Message,
			TTL:         ttl,
		}

		expectedData, _ := json.Marshal(expectedSnapshot)

		mock.ExpectZAdd(fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID), &redis.Z{
			Score:  float64(epochID),
			Member: string(expectedData),
		}).SetVal(1)

		mock.ExpectZRangeByScoreWithScores(fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID), &redis.ZRangeBy{
			Min: "-inf",
			Max: "+inf",
		}).SetVal([]redis.Z{{Score: float64(epochID), Member: string(expectedData)}})

		err := cache.AddUnfinalizedSnapshotCID(ctx, payload, ttl)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Redis error when adding snapshot CID", func(t *testing.T) {
		expectedErr := errors.New("Redis error")

		ttl := time.Now().Unix() + 3600*24
		expectedSnapshot := &datamodel.UnfinalizedSnapshot{
			SnapshotCID: snapshotCID,
			Snapshot:    payload.Message,
			TTL:         ttl,
		}

		expectedData, _ := json.Marshal(expectedSnapshot)

		mock.ExpectZAdd(fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_UNFINALIZED_SNAPSHOT_CIDS, projectID), &redis.Z{
			Score:  float64(epochID),
			Member: string(expectedData),
		}).SetErr(expectedErr)

		err := cache.AddUnfinalizedSnapshotCID(ctx, payload, ttl)
		if err == nil {
			t.Errorf("Expected an error, but no error occurred")
		}
	})

	// Ensure that all expectations were met
	err := mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet Redis expectations: %v", err)
	}
}

func TestRedisCache_AddSnapshotterStatusReport(t *testing.T) {
	mockClient, mock := redismock.NewClientMock()
	cache := RedisCache{
		readClient:  mockClient,
		writeClient: mockClient,
	}

	ctx := context.Background()
	epochID := 1
	projectID := "testProject"
	report := &datamodel.SnapshotterStatusReport{
		SubmittedSnapshotCid: "snapshotCID",
		FinalizedSnapshotCid: "snapshotCID",
		State:                datamodel.MissedSnapshotSubmission,
	}

	t.Run("Add snapshotter status report successfully", func(t *testing.T) {
		reportJSON, _ := json.Marshal(report)

		mock.ExpectHSet(fmt.Sprintf(redisutils.REDIS_KEY_SNAPSHOTTER_STATUS_REPORT, projectID), strconv.Itoa(epochID), string(reportJSON)).SetVal(1)
		mock.ExpectIncr(fmt.Sprintf(redisutils.REDIS_KEY_TOTAL_MISSED_SNAPSHOT_COUNT, projectID)).SetVal(1)

		err := cache.AddSnapshotterStatusReport(ctx, epochID, projectID, report)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Redis error when adding snapshotter status report", func(t *testing.T) {
		reportJSON, _ := json.Marshal(report)

		expectedErr := errors.New("Redis error")

		mock.ExpectHSet(fmt.Sprintf(redisutils.REDIS_KEY_SNAPSHOTTER_STATUS_REPORT, projectID), strconv.Itoa(epochID), string(reportJSON)).SetErr(expectedErr)

		err := cache.AddSnapshotterStatusReport(ctx, epochID, projectID, report)
		if err == nil {
			t.Errorf("Expected an error, but no error occurred")
		}
	})

	t.Run("Redis error when incrementing total missed snapshot count", func(t *testing.T) {
		reportJSON, _ := json.Marshal(report)

		expectedErr := errors.New("Redis error")

		mock.ExpectHSet(fmt.Sprintf(redisutils.REDIS_KEY_SNAPSHOTTER_STATUS_REPORT, projectID), strconv.Itoa(epochID), string(reportJSON)).SetVal(1)
		mock.ExpectIncr(fmt.Sprintf(redisutils.REDIS_KEY_TOTAL_MISSED_SNAPSHOT_COUNT, projectID)).SetErr(expectedErr)

		err := cache.AddSnapshotterStatusReport(ctx, epochID, projectID, report)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Add snapshotter status report with nil report", func(t *testing.T) {
		nilReport := (*datamodel.SnapshotterStatusReport)(nil)

		mock.ExpectIncr(fmt.Sprintf(redisutils.REDIS_KEY_TOTAL_SUCCESSFUL_SNAPSHOT_COUNT, projectID)).SetVal(1)

		err := cache.AddSnapshotterStatusReport(ctx, epochID, projectID, nilReport)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Ensure that all expectations were met
	err := mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet Redis expectations: %v", err)
	}
}

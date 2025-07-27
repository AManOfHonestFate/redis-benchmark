package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	scenarioMaxDuration = time.Second * 10
	tasksPerWorker      = 10000
	workersCount        = 100
	slots               = 1000
)

type BenchmarkScenario interface {
	Run(ctx context.Context, rdb *redis.ClusterClient, tasks int64, workers int) *Statistics
}

func checkClusterHealth(rdb *redis.ClusterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use the first node for cluster info
	info, err := rdb.Do(ctx, "CLUSTER", "INFO").Text()
	if err != nil {
		log.Fatalf("Failed to get cluster info: %v", err)
	}

	log.Printf("CLUSTER INFO:\n%s", info)

	// Parse cluster_state:ok
	stateLine := ""
	for _, line := range splitLines(info) {
		if len(line) > 0 && (line[0] == 'c' || line[0] == 'C') && len(line) > 13 && line[:13] == "cluster_state" {
			stateLine = line
			break
		}
	}
	if stateLine == "" || (len(stateLine) > 0 && !containsOk(stateLine)) {
		log.Fatalf("Cluster is not healthy: %s", stateLine)
	}
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func containsOk(s string) bool {
	return strings.Contains(s, "ok")
}

func main() {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          []string{"localhost:7001", "localhost:7002", "localhost:7003"},
		Protocol:       2,
		Username:       "",
		Password:       "pass",
		RouteByLatency: true,
	})

	err := rdb.ForEachShard(context.Background(), func(ctx context.Context, client *redis.Client) error {
		status, err := client.Ping(ctx).Result()
		if err != nil {
			return err
		}
		log.Printf("shard(%v): %s", client.Options().Addr, status)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	checkClusterHealth(rdb)

	var scenarios []BenchmarkScenario
	scenarios = append(scenarios, &RandomSetsScenario{
		NumberOfSlots: slots,
	})

	for _, scenario := range scenarios {
		stats := scenario.Run(context.Background(), rdb, tasksPerWorker, workersCount)
		stats.Display()
	}
}

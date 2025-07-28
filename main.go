package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	scenarioMaxDuration = time.Second * 10
	tasksPerWorker      = 10000
	workersCount        = 100
	slots               = 10000
)

func main() {
	readEnvFile()

	password, present := os.LookupEnv("REDIS_PASSWORD")
	if !present {
		log.Fatal("REDIS_PASSWORD is not set")
	}

	tcpConnections := atomic.Int64{}
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          []string{"localhost:7001", "localhost:7002", "localhost:7003"},
		Protocol:       2,
		Username:       "",
		Password:       password,
		RouteByLatency: true,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			tcpConnections.Add(1)
			return nil
		},
	})
	defer rdb.Close()

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
	scenarios = append(scenarios,
		&RandomSetsScenario{
			Name: "RandomSets",
		},
		&RandomReadsScenario{
			Name: "RandomReads",
		},
		&RandomReadSetScenario{
			Name: "RandomReadSet",
		},
		&PipelineScenario{
			PipelineSize: 100,
			Name:         "Pipeline",
		},
		&TransactionScenario{
			PipelineSize: 100,
			KeysPerSlot:  10,
			Name:         "Transaction",
		},
	)

	for _, scenario := range scenarios {
		scenario.Setup(rdb, slots)
		stats := scenario.Run(context.Background(), rdb, tasksPerWorker, workersCount)
		stats.TCPConnections = tcpConnections.Load()
		tcpConnections.Store(0)
		log.Printf("\n%s%s\n", scenario.GetName(), stats.Display())
	}
}

type BenchmarkScenario interface {
	Setup(rdb *redis.ClusterClient, slots int)
	Run(ctx context.Context, rdb *redis.ClusterClient, tasks int64, workers int) *Statistics
	GetName() string
}

func checkClusterHealth(rdb *redis.ClusterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use the first node for cluster info
	info, err := rdb.Do(ctx, "CLUSTER", "INFO").Text()
	if err != nil {
		log.Fatalf("Failed to get cluster info: %v", err)
	}

	log.Printf("\n\nCLUSTER INFO:\n%s\n\n", info)

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

func readEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		envVars := strings.Split(scanner.Text(), "=")
		if len(envVars) != 2 {
			log.Fatal("Invalid env file format")
		}
		os.Setenv(envVars[0], envVars[1])
	}
}

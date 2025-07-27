package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

type Statistics struct {
	TCPConnections atomic.Int64
	TotalTasks     atomic.Int64
	ErrorCounter   atomic.Int64
	LatencySum     time.Duration
	TestDuration   time.Duration
	mu             sync.Mutex
}

func (s *Statistics) Display() {
	log.Printf("total tasks: %d, average latency: %v, error: %d, total connections: %d, time: %s",
		s.TotalTasks.Load(), s.LatencySum/time.Duration(s.TotalTasks.Load()), s.ErrorCounter.Load(), s.TCPConnections.Load(), s.TestDuration)
}

type RandomSetsScenario struct {
	wg            sync.WaitGroup
	NumberOfSlots int
}

func SetupContiniousKeys(rdb *redis.Client, slots int) {
	ctx, cansel := context.WithTimeout(context.Background(), time.Second)
	for i := range slots {
		_, err := rdb.Set(ctx, strconv.Itoa(i), i, 0).Result()
		if err != nil {
			cansel()
			log.Fatal("initial set failed", err)
		}
	}
	cansel()
}

func (s *RandomSetsScenario) Run(ctx context.Context, rdb *redis.ClusterClient, tasks int64, workers int) *Statistics {
	stats := Statistics{}
	startTime := time.Now()
	rdb.Options().OnConnect = func(ctx context.Context, cn *redis.Conn) error {
		stats.mu.Lock()
		stats.TCPConnections.Add(1)
		stats.mu.Unlock()
		return nil
	}

	for i := range workers {
		s.wg.Add(1)

		go func() {
			seed := rand.New(rand.NewSource(int64(i)))
			defer s.wg.Done()

			for range tasks {
				task := s.generateTask(seed)

				timePreGet := time.Now()
				_, err := rdb.Set(ctx, strconv.Itoa(task), task, 0).Result()
				latency := time.Since(timePreGet)

				stats.mu.Lock()
				stats.LatencySum += latency
				stats.mu.Unlock()

				if err != nil {
					stats.ErrorCounter.Add(1)
				}
			}
		}()
	}

	s.wg.Wait()

	endTime := time.Now()
	stats.TestDuration = endTime.Sub(startTime)
	stats.TotalTasks.Add(tasks*int64(workers) - stats.ErrorCounter.Load())

	return &stats
}

func (s *RandomSetsScenario) generateTask(generator *rand.Rand) int {
	return generator.Intn(s.NumberOfSlots)
}

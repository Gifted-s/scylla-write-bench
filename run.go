package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
)

// RunConcurrently runs the processRequest function concurrently and uses the wait group
// to indicate the completion of each goroutine. Returns error if concurrency is zero
func RunConcurrently(wg *sync.WaitGroup, processRequest func(threadId int)) error {
	if concurrency == 0 {
		return errors.New("concurrency should be greater than 0")
	}

	for i := 0; i < concurrency; i++ {
		go func(threadId int, wg *sync.WaitGroup) {
			defer wg.Done()
			if tracker.AllowRequest() {
				// process request
				processRequest(threadId)
			} else {
				// throttle request(another option is to queue these requests and
				// run them in the future instead of rejecting them)
				atomic.AddInt64(&requestsThrottled, 1)
			}

		}(i, wg)
	}
	return nil
}

// RunTest exceutes the test function and adds its latency to the overall latency
func RunTest(threadId int, test func() (time.Duration, error)) {
	latency, err := test()
	if err == nil {

		// TODO: We can collect other information about each
		// thread execution and use these to plot a histogram
		updateMetrics(latency)
	} else {
		fmt.Printf("ThreadId: %d  error: %s\n", threadId, err)
		if latency > errorToTimeoutCutoffTime {
			// Consider this error to be timeout error and add its latency to total latency
			updateMetrics(latency)
		}
	}
}

func updateMetrics(latency time.Duration) {
	atomic.AddInt64(&totalLatency, int64(latency))
	atomic.AddInt64(&completedRequests, 1)
}

// DoWrite performs db write operation
func DoWrite(threadId int, session *gocql.Session) {
	RunTest(threadId, func() (time.Duration, error) {
		request := "INSERT INTO " + keyspaceName + "." + tableName + " (key, value) VALUES (?, ?)"
		query := session.Query(request)
		rng := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
		data := generateRandomData(rng, 5, 10)
		bound := query.Bind(data.key, data.value)

		requestStart := time.Now()
		err := bound.Exec()
		requestEnd := time.Now()
		latency := requestEnd.Sub(requestStart)
		if err != nil {
			// Normally we should retry here since failure could be caused by a transient reason.
			// Would discuss this with the team(retry count, policy e.t.c)
			return latency, err
		}

		return latency, nil
	})

}

type randomData struct {
	key   string
	value string
}

func generateRandomData(rng *rand.Rand, keyLength int, valueLength int) *randomData {
	key := randString(rng, keyLength)
	value := randString(rng, valueLength)

	return &randomData{key, value}
}

func randString(rng *rand.Rand, n int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rng.Intn(len(letterBytes))]
	}
	return string(b)
}

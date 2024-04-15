package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocql/gocql"
)

var (
	keyspaceName      string
	tableName         string
	clusterName       string
	replicationFactor int

	concurrency int
	maxRate     int

	processTimeout           time.Duration
	timeout                  time.Duration
	errorToTimeoutCutoffTime time.Duration

	dropDBAfterTest bool

	totalLatency      int64
	completedRequests int64
	reportInterval    time.Duration

	tracker           *WindowTracker
	requestsThrottled int64

	stopAll bool
)

func ExecuteQuery(ctx context.Context, session *gocql.Session, request string) {
	err := session.Query(request).WithContext(ctx).Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func PrepareDatabase(ctx context.Context, session *gocql.Session, replicationFactor int) {
	// NOTE: SimpleStrategy strategy was used here but `NetworkTopologyStrategy` should be used for production case
	ExecuteQuery(ctx, session, fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : %d }", keyspaceName, replicationFactor))
	// Create table and columns
	// keys and values will always be random
	ExecuteQuery(ctx, session, "CREATE TABLE IF NOT EXISTS "+keyspaceName+"."+tableName+" (key text, value text, PRIMARY KEY(key)) WITH compression = { }")
}

func DropDatabase(ctx context.Context, session *gocql.Session) {
	ExecuteQuery(ctx, session, fmt.Sprintf("DROP KEYSPACE  IF EXISTS  %s", keyspaceName))
}

func cleanup(ctx context.Context, session *gocql.Session) {
	defer session.Close()
	if dropDBAfterTest {
		DropDatabase(ctx, session)
	}
}

func main() {
	flag.StringVar(&clusterName, "cluster", "", "cluster name (required)")
	flag.StringVar(&keyspaceName, "keyspace", "scylla_bench", "keyspace name (optional)")
	flag.StringVar(&tableName, "table", "test", "table to use (optional)")

	flag.IntVar(&concurrency, "parallelism", 0, "maximum number of concurrent queries (required)")
	flag.IntVar(&maxRate, "rate-limit", 0, "maximum number of requests per second req/s (required)")

	flag.DurationVar(&processTimeout, "process-timeout", 30*time.Minute, "set how long this process should run before signaling a timeout error")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "each request timeout")

	flag.BoolVar(&dropDBAfterTest, "drop-db", true, "drop database after process is completed")

	flag.DurationVar(&reportInterval, "report-interval", 1*time.Millisecond, "at what interval should total requests performed and the avarage latencies be printed")
	flag.Parse()

	if clusterName == "" {
		log.Fatal("--cluster flag: cluster name must be specified")
	}

	if concurrency == 0 {
		log.Fatal("--parallelism flag: Number of queries to be performed in parallel must be specified")
	}
	if maxRate == 0 {
		log.Fatal("--rate-limit flag: Maximum number of queries to be executed per seconds must be specified")
	}

	cluster := gocql.NewCluster(clusterName)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = timeout
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	replicationFactor = 1 // Should be > 1 for fault tolerance in production mode

	PrepareDatabase(ctx, session, replicationFactor)
	defer cleanup(ctx, session)
	rlconfig := NewRateLimiterConfig(maxRate)
	tracker = NewWindowTracker(rlconfig)

	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)
	go func() {
		<-interrupted
		fmt.Println("\ninterrupted")

		<-interrupted
		fmt.Println("\nkilled")
		cleanup(ctx, session)
		os.Exit(1)
	}()

	go func() {
		time.Sleep(processTimeout)
		fmt.Println("\nProcess took longer than expected, timeout error")
		cleanup(ctx, session)
		os.Exit(1)
	}()

	if timeout != 0 {
		errorToTimeoutCutoffTime = timeout / 5
	} else {
		errorToTimeoutCutoffTime = time.Second
	}

	fmt.Println("Configuration")
	fmt.Println("Concurrency:\t\t", concurrency)
	fmt.Println("Maximum Rate Limit:\t", maxRate, "req/s")
	wg := &sync.WaitGroup{}
	wg.Add(concurrency)

	startTime := time.Now()
	err = RunConcurrently(wg, func(threadId int) {
		DoWrite(threadId, session)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Print total requests and the average latencies at report interval
	go func() {
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()
		for range ticker.C {
			if !stopAll {
				printStat()
			}
		}
	}()

	wg.Wait()
	endTime := time.Now()
	totalTimeTaken := endTime.Sub(startTime)
	printStat()
	fmt.Printf("\nFINAL RESULT:\n")
	fmt.Printf("Requests Sent: \t\t%d\n", concurrency)
	fmt.Printf("Maximum Requests Rate: \t%d/sec\n", maxRate)
	fmt.Printf("Requests Processed: \t%d\n", completedRequests)
	fmt.Printf("Requests Throttled: \t%d\n", requestsThrottled)
	fmt.Printf("Process Execution Time: %v\n", totalTimeTaken)
	stopAll = true
}

func printStat() {
	if atomic.LoadInt64(&completedRequests) == 0 {
		return
	}
	avgLatency := time.Duration(totalLatency / atomic.LoadInt64(&completedRequests))

	fmt.Printf("Total requests: %d, Average latency: %v\n", atomic.LoadInt64(&completedRequests), avgLatency)

}

# scylla-write-bench

scylla-write-bench is a simple CLI for performing write benchmarking for [Scylla](https://github.com/scylladb/scylla) written in Go.
It allows parallel insertion of random data to ScyllaDB while rate-limiting the number of requests. i.e a user can specify the maximum number of parallel writes and the maximum rate at which these writes are performed i.e req/sec. The number of requests performed and average latencies is printed periodically for assessment.

## Install

```
git clone https://github.com/Gifted-s/scylla-write-bench
cd scylla-write-bench/
go build ./
```

#### Flags(required)

- `--cluster` defines cluster name to use e.g <i>localhost:9042</i>.

- `--parallelism` sets maximum number of queries to be performed in parallel.

- `--rate-limit ` sets maximum number of requests to be performed per second.

#### Flags(optional)

- `--keyspace` defines keyspace name to use (default <i>scylla_bench</i>).

- `--table` defines table name to work with (default <i>test</i>).

- `--process-timeout` sets how long this process should run before signaling a timeout error (default <i>30 minutes</i>).

- `--timeout` sets timeout for each request (default <i>5 seconds</i>).

- `--drop-db` drop database after process is completed (default <i>true</i>).

- `--report-interval` sets interval at which total requests performed and the avarage latencies is printed (default <i>1 millisecond</i>).


## Examples

1. Run 1000 queries in parallel allowing only 100 queries per second ```./scylla-benchmarker --cluster="localhost:9042" --parallelism=1000 --rate-limit=100```


2. Run 3000 queries in parallel allowing only 500 queries per second ```./scylla-benchmarker --cluster="localhost:9042" --parallelism=3000 --rate-limit=500```


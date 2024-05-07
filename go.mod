module scylla-benchmarker

go 1.20

replace github.com/gocql/gocql => github.com/scylladb/gocql v1.13.0

require github.com/gocql/gocql v1.2.1

require (
	github.com/golang/snappy v0.0.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/pkg/errors v0.9.1
	gopkg.in/inf.v0 v0.9.1 // indirect
)

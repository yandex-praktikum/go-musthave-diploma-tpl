package config

import "time"

const (
	DefaultDomain           = "localhost:8080"
	DefaultClientDomain     = "localhost:8081"
	DefaultLoggerLevel      = "debug"
	DefaultDatabaseDsn      = "host=127.0.0.1 user='username' password='userpassword' dbname='market' sslmode=disable"
	DefaultContextTimeout   = time.Minute
	DefaultReadableTicket   = 10
	DefaultOwner            = 42
	DefaultBatchSize        = 1
	DefaultClientConnection = 2
)

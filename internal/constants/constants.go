package constants

import "time"

const (
	LogLevel   = "info"
	DevMode    = true
	Type       = "plaintext"
	HashHeader = "Authorization"
	UserCtx    = "userId"
	TokenTTL   = 12 * time.Hour
	SigningKey = "kljksj542ds;flks;l;"

	RetryMax     int           = 3
	RetryWaitMin time.Duration = 1 * time.Second
	RetryMedium  time.Duration = 3 * time.Second
	RetryWaitMax time.Duration = 5 * time.Second
)

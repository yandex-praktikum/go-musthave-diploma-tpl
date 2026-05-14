package handler

type Config struct {
	EncryptionKey string `yaml:"EncryptionKey"`

	ExpirationTimePerMinute int `yaml:"ExpirationTimePerMinute"`
}

package service

type Config struct {
	SaltLen uint32 `yaml:"SaltLen"`

	KeyLength uint32 `yaml:"KeyLength"`
}

package config

import (
	"testing"
)

func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		SelfAddress:          ":8080",
		DSN:                  "postgres://localhost/db",
		AccrualSystemAddress: "http://localhost:8081",
		ENV:                  "test",
		PrivateKeyJWT:        "secret",
		PrivateKey:           "key",
	}

	if cfg.SelfAddress != ":8080" {
		t.Errorf("SelfAddress = %v, want :8080", cfg.SelfAddress)
	}
	if cfg.DSN != "postgres://localhost/db" {
		t.Errorf("DSN = %v, want postgres://localhost/db", cfg.DSN)
	}
	if cfg.AccrualSystemAddress != "http://localhost:8081" {
		t.Errorf("AccrualSystemAddress = %v, want http://localhost:8081", cfg.AccrualSystemAddress)
	}
	if cfg.ENV != "test" {
		t.Errorf("ENV = %v, want test", cfg.ENV)
	}
	if cfg.PrivateKeyJWT != "secret" {
		t.Errorf("PrivateKeyJWT = %v, want secret", cfg.PrivateKeyJWT)
	}
	if cfg.PrivateKey != "key" {
		t.Errorf("PrivateKey = %v, want key", cfg.PrivateKey)
	}
}

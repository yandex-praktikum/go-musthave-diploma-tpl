package config

import (
	"reflect"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse(nil, emptyEnv)
	if err != nil {
		t.Fatal(err)
	}

	want := Config{
		RunAddress:           ":8080",
		DatabaseURI:          "",
		AccrualSystemAddress: "",
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("config mismatch: got %+v want %+v", cfg, want)
	}
}

func TestParseFlags(t *testing.T) {
	cfg, err := Parse([]string{"-a", ":9090", "-d", "postgres://flags", "-r", "http://accrual"}, emptyEnv)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.RunAddress != ":9090" || cfg.DatabaseURI != "postgres://flags" || cfg.AccrualSystemAddress != "http://accrual" {
		t.Fatalf("flags were not applied: %+v", cfg)
	}
}

func TestParseEnv(t *testing.T) {
	cfg, err := Parse(nil, mapEnv(map[string]string{
		"RUN_ADDRESS":            ":7070",
		"DATABASE_URI":           "postgres://env",
		"ACCRUAL_SYSTEM_ADDRESS": "http://env-accrual",
	}))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.RunAddress != ":7070" || cfg.DatabaseURI != "postgres://env" || cfg.AccrualSystemAddress != "http://env-accrual" {
		t.Fatalf("env was not applied: %+v", cfg)
	}
}

func TestParseEnvOverridesFlags(t *testing.T) {
	cfg, err := Parse([]string{"-a", ":9090", "-d", "postgres://flags", "-r", "http://flags-accrual"}, mapEnv(map[string]string{
		"RUN_ADDRESS":            ":7070",
		"DATABASE_URI":           "postgres://env",
		"ACCRUAL_SYSTEM_ADDRESS": "http://env-accrual",
	}))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.RunAddress != ":7070" || cfg.DatabaseURI != "postgres://env" || cfg.AccrualSystemAddress != "http://env-accrual" {
		t.Fatalf("env did not override flags: %+v", cfg)
	}
}

func emptyEnv(string) (string, bool) {
	return "", false
}

func mapEnv(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

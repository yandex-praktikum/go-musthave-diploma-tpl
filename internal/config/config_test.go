package config

import (
	"testing"
)

const (
	SelfAddress          = ":8081"
	DSN                  = "postgres://localhost/db"
	AccrualSystemAddress = "http://localhost:8081"
	ENV                  = "release"
	PrivateKeyJWT        = "default-key"
	PrivateKey           = "6ba7b810-9dad-11d1-80b4-00b74fd430c8"
)

func TestConfigStruct(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    Config
	}{
		{
			name: "All environments setup",
			envVars: map[string]string{
				"RUN_ADDRESS":            SelfAddress,
				"DATABASE_URI":           DSN,
				"ACCRUAL_SYSTEM_ADDRESS": AccrualSystemAddress,
				"ENV":                    ENV,
				"PRIVATE_KEY_JWT":        PrivateKeyJWT,
				"PRIVATE_KEY":            PrivateKey,
			},
			want: Config{
				SelfAddress:          SelfAddress,
				DSN:                  DSN,
				AccrualSystemAddress: AccrualSystemAddress,
				ENV:                  ENV,
				PrivateKeyJWT:        PrivateKeyJWT,
				PrivateKey:           PrivateKey,
			},
		},
		{
			name: "Missing server address",
			envVars: map[string]string{
				"DATABASE_URI":           DSN,
				"ACCRUAL_SYSTEM_ADDRESS": AccrualSystemAddress,
				"ENV":                    ENV,
				"PRIVATE_KEY_JWT":        PrivateKeyJWT,
				"PRIVATE_KEY":            PrivateKey,
			},
			want: Config{
				SelfAddress:          ":8080",
				DSN:                  DSN,
				AccrualSystemAddress: AccrualSystemAddress,
				ENV:                  ENV,
				PrivateKeyJWT:        PrivateKeyJWT,
				PrivateKey:           PrivateKey,
			},
		},
		{
			name: "Empty environments - use flags",
			envVars: map[string]string{
				"ENV":             ENV,
				"PRIVATE_KEY_JWT": PrivateKeyJWT,
				"PRIVATE_KEY":     PrivateKey,
			},
			want: Config{
				SelfAddress:          ":8080",
				DSN:                  "",
				AccrualSystemAddress: "",
				ENV:                  ENV,
				PrivateKeyJWT:        PrivateKeyJWT,
				PrivateKey:           PrivateKey,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for key, val := range tt.envVars {
				t.Setenv(key, val)
			}

			got, _ := New()

			if got.SelfAddress != tt.want.SelfAddress {
				t.Errorf("RUN_ADDRESS = %v, want %v", got.SelfAddress, tt.want.SelfAddress)
			}
			if got.DSN != tt.want.DSN {
				t.Errorf("DATABASE_URI = %v, want %v", got.DSN, tt.want.DSN)
			}
			if got.AccrualSystemAddress != tt.want.AccrualSystemAddress {
				t.Errorf("ACCRUAL_SYSTEM_ADDRESS = %v, want %v", got.AccrualSystemAddress, tt.want.AccrualSystemAddress)
			}
			if got.ENV != tt.want.ENV {
				t.Errorf("ENV = %v, want %v", got.ENV, tt.want.ENV)
			}
			if got.PrivateKeyJWT != tt.want.PrivateKeyJWT {
				t.Errorf("PRIVATE_KEY_JWT = %v, want %v", got.PrivateKeyJWT, tt.want.PrivateKeyJWT)
			}
			if got.PrivateKey != tt.want.PrivateKey {
				t.Errorf("PRIVATE_KEY = %v, want %v", got.PrivateKey, tt.want.PrivateKey)
			}
		})
	}

}

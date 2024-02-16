package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type Flags struct {
	FlagAddr        string
	FlagDBAddr      string
	FlagAccuralAddr string
}

type ServerENV struct {
	Address     string `env:"RUN_ADDRESS"`
	DBAddress   string `env:"DATABASE_URI"`
	AccuralAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func ShaData(result string, key string) string {
	b := []byte(result)
	shakey := []byte(key)
	// создаём новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := hmac.New(sha256.New, shakey)
	// передаём байты для хеширования
	h.Write(b)
	// вычисляем хеш
	hash := h.Sum(nil)
	sha := base64.URLEncoding.EncodeToString(hash)
	return string(sha)
}

func ParseFlagsAndENV() Flags {
	var Flag Flags
	flag.StringVar(&Flag.FlagAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Flag.FlagDBAddr, "d", "", "address for db")
	flag.StringVar(&Flag.FlagAccuralAddr, "r", "", "accural system addr")
	flag.Parse()
	var envcfg ServerENV
	err := env.Parse(&envcfg)
	if err != nil {
		log.Fatal(err)
	}

	if len(envcfg.Address) > 0 {
		Flag.FlagAddr = envcfg.Address
	}
	if len(envcfg.DBAddress) > 0 {
		Flag.FlagDBAddr = envcfg.DBAddress
	}

	if len(envcfg.AccuralAddr) > 0 {
		Flag.FlagAccuralAddr = envcfg.AccuralAddr
	}

	return Flag
}

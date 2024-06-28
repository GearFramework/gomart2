package config

import (
	"os"
)

type GomartConfig struct {
	Addr        string
	DatabaseDSN string
	AccrualAddr string
}

func GetConfig() *GomartConfig {
	conf := ParseFlags()
	if envAddr := os.Getenv("RUN_ADDRESS"); envAddr != "" {
		conf.Addr = envAddr
	}
	if envDatabaseDSN := os.Getenv("DATABASE_URI"); envDatabaseDSN != "" {
		conf.DatabaseDSN = envDatabaseDSN
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		conf.AccrualAddr = envAccrualAddr
	}
	return conf
}

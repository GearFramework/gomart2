package config

import (
	"flag"
)

const (
	defaultAddress     = ":8080"
	defaultDatabaseDSN = ""
)

func ParseFlags() *GomartConfig {
	var conf GomartConfig
	flag.StringVar(&conf.Addr, "a", defaultAddress, "address to run server")
	flag.StringVar(&conf.DatabaseDSN, "d", defaultDatabaseDSN, "database connection DSN")
	flag.StringVar(&conf.AccrualAddr, "r", "", "address of accrual system")
	flag.Parse()
	return &conf
}

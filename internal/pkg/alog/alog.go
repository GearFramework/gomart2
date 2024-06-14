package alog

import (
	"fmt"
	"go.uber.org/zap"
)

func NewLogger(level string) *zap.SugaredLogger {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		lvl = zap.NewAtomicLevel()
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		fmt.Println(err.Error())
	}
	return zl.Sugar()
}

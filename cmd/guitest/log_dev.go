//go:build dev

package main

import (
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

func newLogger() (logger *slog.Logger, sync func() error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapL := zap.Must(config.Build())
	return slog.New(zapslog.NewHandler(zapL.Core(), &zapslog.HandlerOptions{
		AddSource: true,
	})), zapL.Sync
}

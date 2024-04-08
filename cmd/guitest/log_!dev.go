//go:build !dev

package main

import (
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

func newLogger() (logger *slog.Logger, sync func() error) {
	config := zap.NewProductionConfig()
	zapL := zap.Must(config.Build())
	return slog.New(zapslog.NewHandler(zapL.Core(), nil)), zapL.Sync
}

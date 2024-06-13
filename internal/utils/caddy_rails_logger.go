package utils

import (
	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CaddyRailsLogger struct {
	logger *zap.Logger
	prefix string
}

const DefaultPrefix = "[CaddyRails] "

func NewCaddyRailsLogger() *CaddyRailsLogger {
	return &CaddyRailsLogger{
		logger: caddy.Log(),
		prefix: DefaultPrefix,
	}
}

func (pl *CaddyRailsLogger) Info(msg string, fields ...zapcore.Field) {
	pl.logger.Info(pl.prefix+msg, fields...)
}

func (pl *CaddyRailsLogger) Error(msg string, fields ...zapcore.Field) {
	pl.logger.Error(pl.prefix+msg, fields...)
}

func (pl *CaddyRailsLogger) Warn(msg string, fields ...zapcore.Field) {
	pl.logger.Warn(pl.prefix+msg, fields...)
}

func (pl *CaddyRailsLogger) Debug(msg string, fields ...zapcore.Field) {
	pl.logger.Debug(pl.prefix+msg, fields...)
}

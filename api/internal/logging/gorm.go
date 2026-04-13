package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	logLevel                  gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

func NewGormLogger(appLevel slog.Level) gormlogger.Interface {
	var gormLevel gormlogger.LogLevel
	switch {
	case appLevel <= slog.LevelDebug:
		gormLevel = gormlogger.Warn
	default:
		gormLevel = gormlogger.Error
	}

	return &GormLogger{
		logLevel:                  gormLevel,
		slowThreshold:             200 * time.Millisecond,
		ignoreRecordNotFoundError: true,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.logLevel = level
	return &clone
}

func (l *GormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Info {
		return
	}

	slog.Info("gorm info", "component", "database", "message", fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Warn {
		return
	}

	slog.Warn("gorm warning", "component", "database", "message", fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Error {
		return
	}

	slog.Error("gorm error", "component", "database", "message", fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []any{
		"component", "database",
		"elapsed_ms", float64(elapsed.Nanoseconds()) / 1e6,
		"rows", rows,
		"sql", sql,
	}

	switch {
	case err != nil && l.logLevel >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		slog.Error("database query failed", append(fields, "error", err.Error())...)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.logLevel >= gormlogger.Warn:
		slog.Warn("slow database query", append(fields, "slow_threshold_ms", float64(l.slowThreshold.Nanoseconds())/1e6)...)
	case l.logLevel >= gormlogger.Info:
		slog.Debug("database query executed", fields...)
	}
}

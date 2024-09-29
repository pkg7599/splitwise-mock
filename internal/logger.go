package internal

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormLogger "gorm.io/gorm/logger"
)

type CustomLogger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
}

func Logger() CustomLogger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	logger.Debug("Logger Initialized")
	return logger
}

var Log CustomLogger = Logger()

type ZapLogger struct {
	logger *zap.Logger
}

func (z ZapLogger) LogMode(gormLogger.LogLevel) gormLogger.Interface {
	return z
}

func (z ZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Info(msg, zap.Any("data", data))
}

func (z ZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Warn(msg, zap.Any("data", data))
}

func (z ZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Error(msg, zap.Any("data", data))
}

func (z ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	msg := "trace:"
	if err != nil {
		msg = err.Error()
	}
	z.logger.Debug(msg, zap.Any("begin", begin), zap.Any("sql", sql), zap.Any("rowsAffected", rowsAffected))
}

func GormLogger() ZapLogger {
	zapLogger := Log.(*zap.Logger)
	logger := ZapLogger{logger: zapLogger}
	logger.LogMode(gormLogger.Info)
	return logger
}

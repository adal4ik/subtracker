package logger

import (
	"go.uber.org/zap"
)

const (
	EnvProd = "production"
	EnvDev  = "development"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Sync() error
}

type zapLogger struct {
	logger *zap.Logger
}

func New(env string) Logger {
	var cfg zap.Config

	switch env {
	case EnvProd:
		cfg = zap.NewProductionConfig()
	case EnvDev:
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.DisableStacktrace = true

	logger, err := cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic("cannot initialize zap logger: " + err.Error())
	}

	return &zapLogger{logger: logger}
}

func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func NewNopLogger() Logger {
	return &zapLogger{logger: zap.NewNop()}
}
